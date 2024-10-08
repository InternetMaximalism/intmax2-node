use crate::{
    app::{
        encode::decode_plonky2_proof,
        interface::{
            ProofResponse, ProofSendRequest, ProofSendValue, ProofsSendResponse, SendIdQuery,
        },
        state::AppState,
    },
    proof::{generate_balance_send_proof_job, RedisResponse},
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use intmax2_zkp::{
    common::witness::update_witness::UpdateWitness,
    constants::NUM_TRANSFERS_IN_TX,
    ethereum_types::{u256::U256, u32limb_trait::U32LimbTrait},
};

#[get("/proof/{public_key}/send/{request_id}")]
async fn get_proof(
    query_params: web::Path<(String, String)>,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let public_key = U256::from_hex(&query_params.0).expect("failed to parse public key");
    let request_id = &query_params.1;
    let proof_json = redis::Cmd::get(&get_balance_send_request_id(
        &public_key.to_hex(),
        request_id,
    ))
    .query_async::<_, Option<String>>(&mut conn)
    .await
    .map_err(error::ErrorInternalServerError)?;

    match proof_json {
        Some(proof_json) => {
            let proof: RedisResponse =
                serde_json::from_str(&proof_json).map_err(error::ErrorInternalServerError)?;

            if proof.success {
                let response = ProofResponse {
                    success: true,
                    request_id: request_id.clone(),
                    proof: Some(proof.message),
                    error_message: None,
                };

                Ok(HttpResponse::Ok().json(response))
            } else {
                let response = ProofResponse {
                    success: false,
                    request_id: request_id.clone(),
                    proof: None,
                    error_message: Some(proof.message),
                };

                Ok(HttpResponse::Ok().json(response))
            }
        }
        None => {
            let response = ProofResponse {
                success: false,
                request_id: request_id.clone(),
                proof: None,
                error_message: Some("balance proof is not generated".to_string()),
            };

            Ok(HttpResponse::Ok().json(response))
        }
    }
}

#[get("/proofs/{public_key}/send")]
async fn get_proofs(
    query_params: web::Path<String>,
    req: HttpRequest,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder, actix_web::Error> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let public_key = U256::from_hex(&query_params).expect("failed to parse public key");

    let query_string = req.query_string();
    let ids_query = serde_qs::from_str::<SendIdQuery>(query_string);

    let request_ids: Vec<String> = match ids_query {
        Ok(query) => query.request_ids,
        Err(e) => {
            log::warn!("Failed to deserialize query: {:?}", e);
            return Ok(HttpResponse::BadRequest().body("Invalid query parameters"));
        }
    };

    let mut proofs: Vec<ProofSendValue> = Vec::new();
    for request_id in &request_ids {
        let proof_json = redis::Cmd::get(&get_balance_send_request_id(
            &public_key.to_hex(),
            request_id,
        ))
        .query_async::<_, Option<String>>(&mut conn)
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;
        let some_proof = match proof_json {
            Some(proof_json) => {
                let proof: RedisResponse =
                    serde_json::from_str(&proof_json).map_err(error::ErrorInternalServerError)?;

                if proof.success {
                    Some(proof.message)
                } else {
                    let response = ProofsSendResponse {
                        success: false,
                        proofs,
                        error_message: Some(proof.message),
                    };

                    return Ok(HttpResponse::Ok().json(response));
                }
            }
            None => None,
        };
        if let Some(proof) = some_proof {
            proofs.push(ProofSendValue {
                request_id: (*request_id).to_string(),
                proof,
            });
        }
    }

    let response = ProofsSendResponse {
        success: true,
        proofs,
        error_message: None,
    };

    Ok(HttpResponse::Ok().json(response))
}

#[post("/proof/{public_key}/send")]
async fn generate_proof(
    query_params: web::Path<String>,
    req: web::Json<ProofSendRequest>,
    redis: web::Data<redis::Client>,
    state: web::Data<AppState>,
) -> Result<impl Responder> {
    let mut redis_conn = redis
        .get_async_connection()
        .await
        .map_err(error::ErrorInternalServerError)?;

    let public_key = U256::from_hex(&query_params).expect("failed to parse public key");

    let validity_circuit_data = state
        .validity_circuit
        .get()
        .ok_or_else(|| error::ErrorInternalServerError("validity circuit not initialized"))?
        .data
        .verifier_data();
    let encoded_validity_proof = req.balance_update_witness.validity_proof.clone();
    let validity_proof = decode_plonky2_proof(&encoded_validity_proof, &validity_circuit_data)
        .map_err(error::ErrorInternalServerError)?;
    validity_circuit_data
        .verify(validity_proof.clone())
        .map_err(error::ErrorInternalServerError)?;
    // let validity_public_inputs = ValidityPublicInputs::from_pis(&validity_proof.public_inputs);
    let balance_update_witness = UpdateWitness {
        validity_proof,
        block_merkle_proof: req.balance_update_witness.block_merkle_proof.clone(),
        account_membership_proof: req.balance_update_witness.account_membership_proof.clone(),
    };

    // let request_id = validity_public_inputs.public_state.block_hash.to_hex();
    let request_id = req.request_id.clone();
    let full_request_id = get_balance_send_request_id(&public_key.to_hex(), &request_id);
    log::debug!("request ID: {:?}", full_request_id);
    let old_proof = redis::Cmd::get(&full_request_id)
        .query_async::<_, Option<String>>(&mut redis_conn)
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;
    if let Some(old_proof) = old_proof {
        let response = ProofResponse {
            success: true,
            request_id: request_id.clone(),
            proof: Some(old_proof),
            error_message: Some("balance proof already requested".to_string()),
        };

        return Ok(HttpResponse::Ok().json(response));
    }

    let balance_circuit_data = state
        .balance_processor
        .get()
        .ok_or_else(|| error::ErrorInternalServerError("balance processor not initialized"))?
        .balance_circuit
        .data
        .verifier_data();
    let prev_balance_proof = if let Some(req_prev_balance_proof) = &req.prev_balance_proof {
        log::debug!("requested proof size: {}", req_prev_balance_proof.len());
        let prev_balance_proof =
            decode_plonky2_proof(&req_prev_balance_proof, &balance_circuit_data)
                .map_err(error::ErrorInternalServerError)?;
        balance_circuit_data
            .verify(prev_balance_proof.clone())
            .map_err(error::ErrorInternalServerError)?;

        Some(prev_balance_proof)
    } else {
        None
    };

    // Validation check of balance_witness
    let send_witness = req.send_witness.clone();
    if send_witness.transfers.len() != NUM_TRANSFERS_IN_TX {
        println!(
            "Invalid number of transfers: {}",
            send_witness.transfers.len()
        );
        return Err(error::ErrorBadRequest("Invalid number of transfers"));
    }
    if send_witness.prev_balances.len() != NUM_TRANSFERS_IN_TX {
        println!(
            "Invalid number of prev_balances: {}",
            send_witness.prev_balances.len()
        );
        return Err(error::ErrorBadRequest("Invalid number of prev_balances"));
    }
    if send_witness.asset_merkle_proofs.len() != NUM_TRANSFERS_IN_TX {
        println!(
            "Invalid number of asset_merkle_proofs: {}",
            send_witness.asset_merkle_proofs.len()
        );
        return Err(error::ErrorBadRequest(
            "Invalid number of asset_merkle_proofs",
        ));
    }

    let response = ProofResponse {
        success: true,
        request_id: request_id.clone(),
        proof: None,
        error_message: Some(format!(
            "balance proof (request ID: {request_id}) is generating",
        )),
    };

    // Spawn a new task to generate the proof
    actix_web::rt::spawn(async move {
        let response = generate_balance_send_proof_job(
            full_request_id,
            public_key,
            prev_balance_proof,
            &send_witness,
            &balance_update_witness,
            state
                .balance_processor
                .get()
                .expect("balance processor not initialized"),
            state
                .validity_circuit
                .get()
                .expect("validity circuit not initialized"),
            &mut redis_conn,
        )
        .await;

        match response {
            Ok(v) => {
                log::info!("Proof generation completed (request ID: {request_id})");
                Ok(v)
            }
            Err(e) => {
                log::error!("Failed to generate proof: {:?}", e);
                Err(e)
            }
        }
    });

    Ok(HttpResponse::Ok().json(response))
}

fn get_balance_send_request_id(public_key: &str, block_hash: &str) -> String {
    format!("balance-validity/{}/send/{}", public_key, block_hash)
}
