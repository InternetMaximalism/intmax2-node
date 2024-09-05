use crate::{
    app::{
        encode::decode_plonky2_proof,
        interface::{
            ProofResponse, ProofTransferRequest, ProofTransferValue, ProofsTransferResponse,
            TransferIdQuery,
        },
        state::AppState,
    },
    proof::generate_balance_transfer_proof_job,
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use intmax2_zkp::{
    circuits::balance::balance_pis::BalancePublicInputs,
    ethereum_types::{u256::U256, u32limb_trait::U32LimbTrait},
};

#[get("/proof/{public_key}/transfer/{private_commitment}")]
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
    let proof = redis::Cmd::get(&get_balance_transfer_request_id(
        &public_key.to_hex(),
        request_id,
    ))
    .query_async::<_, Option<String>>(&mut conn)
    .await
    .map_err(error::ErrorInternalServerError)?;

    if proof.is_none() {
        let response = ProofResponse {
            success: false,
            request_id: request_id.clone(),
            proof: None,
            error_message: Some(format!(
                "balance proof is not generated (private_commitment: {request_id})",
            )),
        };

        return Ok(HttpResponse::Ok().json(response));
    }

    let response = ProofResponse {
        success: true,
        request_id: request_id.clone(),
        proof,
        error_message: None,
    };

    Ok(HttpResponse::Ok().json(response))
}

#[get("/proofs/{public_key}/transfer")]
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
    let ids_query = serde_qs::from_str::<TransferIdQuery>(query_string);
    let private_commitments: Vec<String>;

    match ids_query {
        Ok(query) => {
            private_commitments = query.private_commitments;
        }
        Err(e) => {
            log::warn!("Failed to deserialize query: {:?}", e);
            return Ok(HttpResponse::BadRequest().body("Invalid query parameters"));
        }
    }

    let mut proofs: Vec<ProofTransferValue> = Vec::new();
    for private_commitment in &private_commitments {
        let some_proof = redis::Cmd::get(&get_balance_transfer_request_id(
            &public_key.to_hex(),
            private_commitment,
        ))
        .query_async::<_, Option<String>>(&mut conn)
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;
        if let Some(proof) = some_proof {
            proofs.push(ProofTransferValue {
                private_commitment: (*private_commitment).to_string(),
                proof,
            });
        }
    }

    let response = ProofsTransferResponse {
        success: true,
        proofs,
        error_message: None,
    };

    Ok(HttpResponse::Ok().json(response))
}

#[post("/proof/{public_key}/transfer")]
async fn generate_proof(
    query_params: web::Path<String>,
    req: web::Json<ProofTransferRequest>,
    redis: web::Data<redis::Client>,
    state: web::Data<AppState>,
) -> Result<impl Responder> {
    let mut redis_conn = redis
        .get_async_connection()
        .await
        .map_err(error::ErrorInternalServerError)?;

    let public_key = U256::from_hex(&query_params).expect("failed to parse public key");

    let balance_circuit_data = state
        .balance_processor
        .get()
        .ok_or_else(|| error::ErrorInternalServerError("validity circuit not initialized"))?
        .balance_circuit
        .data
        .verifier_data();

    let receive_transfer_witness = req
        .receive_transfer_witness
        .decode(&balance_circuit_data)
        .map_err(error::ErrorInternalServerError)?;
    balance_circuit_data
        .verify(receive_transfer_witness.balance_proof.clone())
        .map_err(error::ErrorInternalServerError)?;
    let balance_public_inputs =
        BalancePublicInputs::from_pis(&receive_transfer_witness.balance_proof.public_inputs);

    // let block_hash = balance_public_inputs.public_state.block_hash;
    let request_id = balance_public_inputs.private_commitment.to_string();
    let full_request_id = get_balance_transfer_request_id(&public_key.to_hex(), &request_id);
    log::debug!("request ID: {:?}", full_request_id);
    let old_proof = redis::Cmd::get(&full_request_id)
        .query_async::<_, Option<String>>(&mut redis_conn)
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;
    if let Some(old_proof) = old_proof {
        let response = ProofResponse {
            success: true,
            request_id,
            proof: Some(old_proof),
            error_message: Some("balance proof already requested".to_string()),
        };

        return Ok(HttpResponse::Ok().json(response));
    }

    let prev_balance_proof = if let Some(req_prev_balance_proof) = &req.prev_balance_proof {
        log::debug!("requested proof size: {}", req_prev_balance_proof.len());
        let prev_balance_proof =
            decode_plonky2_proof(req_prev_balance_proof, &balance_circuit_data)
                .map_err(error::ErrorInternalServerError)?;
        balance_circuit_data
            .verify(prev_balance_proof.clone())
            .map_err(error::ErrorInternalServerError)?;

        Some(prev_balance_proof)
    } else {
        None
    };

    // TODO: Validation check of balance_witness

    let response = ProofResponse {
        success: true,
        request_id: request_id.clone(),
        proof: None,
        error_message: Some(format!(
            "balance proof (request ID: {}) is generating",
            request_id
        )),
    };

    // Spawn a new task to generate the proof
    actix_web::rt::spawn(async move {
        let response = generate_balance_transfer_proof_job(
            full_request_id,
            public_key,
            prev_balance_proof,
            &receive_transfer_witness,
            state
                .balance_processor
                .get()
                .expect("balance processor not initialized"),
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

fn get_balance_transfer_request_id(public_key: &str, private_commitment: &str) -> String {
    format!(
        "balance-validity/{}/transfer/{}",
        public_key, private_commitment
    )
}
