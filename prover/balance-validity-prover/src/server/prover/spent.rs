use crate::{
    app::{
        encode::decode_plonky2_proof,
        interface::{
            ErrorResponse, ProofResponse, ProofSpentRequest, ProofSpentValue, ProofsSpentResponse,
            SpentIdQuery,
        },
        state::AppState,
    },
    proof::generate_balance_single_send_proof_job,
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use intmax2_zkp::{
    circuits::validity::validity_pis::ValidityPublicInputs,
    common::witness::update_witness::UpdateWitness, constants::NUM_TRANSFERS_IN_TX,
};

#[get("/proof/spent/{request_id}")]
async fn get_proof(
    query_params: web::Path<(String, String)>,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let request_id = &query_params.1;
    let proof = redis::Cmd::get(&get_balance_spent_request_id(request_id))
        .query_async::<_, Option<String>>(&mut conn)
        .await
        .map_err(error::ErrorInternalServerError)?;

    if proof.is_none() {
        let response = ProofResponse {
            success: false,
            request_id: request_id.clone(),
            proof: None,
            error_message: Some(format!(
                "balance proof is not generated (request ID: {request_id})",
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

#[get("/proofs/spent")]
async fn get_proofs(
    req: HttpRequest,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder, actix_web::Error> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let query_string = req.query_string();
    let ids_query = serde_qs::from_str::<SpentIdQuery>(query_string);

    let request_ids = match ids_query {
        Ok(query) => query.request_id,
        Err(e) => {
            log::warn!("Failed to deserialize query: {:?}", e);
            return Ok(HttpResponse::BadRequest().body("Invalid query parameters"));
        }
    };

    let mut proofs: Vec<ProofSpentValue> = Vec::new();
    for request_id in &request_ids {
        let some_proof = redis::Cmd::get(&get_balance_spent_request_id(request_id))
            .query_async::<_, Option<String>>(&mut conn)
            .await
            .map_err(actix_web::error::ErrorInternalServerError)?;
        if let Some(proof) = some_proof {
            proofs.push(ProofSpentValue {
                request_id: (*request_id).to_string(),
                proof,
            });
        }
    }

    let response = ProofsSpentResponse {
        success: true,
        proofs,
        error_message: None,
    };

    Ok(HttpResponse::Ok().json(response))
}

#[post("/proof/spent")]
async fn generate_proof(
    req: web::Json<ProofSpentRequest>,
    redis: web::Data<redis::Client>,
    state: web::Data<AppState>,
) -> Result<impl Responder> {
    let mut redis_conn = redis
        .get_async_connection()
        .await
        .map_err(error::ErrorInternalServerError)?;

    let request_id = uuid::Uuid::new_v4();
    let full_request_id = get_balance_spent_request_id(&request_id.to_string());
    log::debug!("request ID: {:?}", full_request_id);
    let old_proof = redis::Cmd::get(&full_request_id)
        .query_async::<_, Option<String>>(&mut redis_conn)
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;
    if let Some(old_proof) = old_proof {
        let response = ProofResponse {
            success: true,
            request_id: request_id.to_string(),
            proof: Some(old_proof),
            error_message: Some("balance proof already requested".to_string()),
        };

        return Ok(HttpResponse::Ok().json(response));
    }

    let instant = std::time::Instant::now();

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
        return Err(error::ErrorBadRequest("Invalid number of transfers"));
    }
    if send_witness.asset_merkle_proofs.len() != NUM_TRANSFERS_IN_TX {
        println!(
            "Invalid number of asset_merkle_proofs: {}",
            send_witness.asset_merkle_proofs.len()
        );
        return Err(error::ErrorBadRequest("Invalid number of transfers"));
    }

    let validity_pis =
        ValidityPublicInputs::from_pis(&balance_update_witness.validity_proof.public_inputs);
    if validity_pis != send_witness.tx_witness.validity_pis {
        return Err(error::ErrorBadRequest("validity proof pis mismatch"));
    }

    let response = generate_balance_single_send_proof_job(
        &send_witness,
        &balance_update_witness,
        state
            .balance_processor
            .get()
            .expect("balance processor not initialized"),
        state
            .validity_circuit
            .get()
            .expect("balance processor not initialized"),
    );
    // let response = generate_balance_spent_proof_job(
    //     &send_witness,
    //     state
    //         .balance_processor
    //         .get()
    //         .expect("balance processor not initialized"),
    // );

    // // Spawn a new task to generate the proof
    // actix_web::rt::spawn(async move {
    //     let response = generate_balance_spent_proof_job(
    //         full_request_id,
    //         &send_witness,
    //         state
    //             .balance_processor
    //             .get()
    //             .expect("balance processor not initialized"),
    //         &mut redis_conn,
    //     );
    //     match response {
    //         Ok(proof) => {
    //             let opts = SetOptions::default()
    //                 .conditional_set(ExistenceCheck::NX)
    //                 .get(true)
    //                 .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

    //             let _ = redis::Cmd::set_options(
    //                 &full_request_id,
    //                 proof.clone(),
    //                 opts,
    //             )
    //             .query_async::<_, Option<String>>(conn)
    //             .await
    //             .with_context(|| "Failed to set proof")?;
    //             log::info!("Proof generation completed (request ID: {request_id})");
    //             Ok(())
    //         }
    //         Err(e) => {
    //             log::error!("Failed to generate proof: {:?}", e);
    //             Err(e)
    //         }
    //     }
    // });

    match response {
        Ok(proof) => {
            log::info!("Proof generation completed (request ID: {request_id})");
            let response = ProofResponse {
                success: true,
                request_id: request_id.to_string(),
                proof: Some(proof),
                error_message: Some(format!(
                    "balance proof (request ID: {request_id}) is generating",
                )),
            };
            println!("Proving time: {:?}", instant.elapsed());

            Ok(HttpResponse::Ok().json(response))
        }
        Err(e) => {
            log::error!("Failed to generate proof: {:?}", e);
            let response = ErrorResponse {
                success: false,
                code: 500,
                message: format!("Failed to generate proof: {e:?}"),
            };

            Ok(HttpResponse::Ok().json(response))
        }
    }
}

fn get_balance_spent_request_id(request_id: &str) -> String {
    format!("balance-validity/spent/{}", request_id)
}
