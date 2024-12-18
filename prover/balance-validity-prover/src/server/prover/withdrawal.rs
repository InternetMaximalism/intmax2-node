use crate::{
    app::{
        encode::decode_plonky2_proof,
        interface::{
            ErrorResponse, ProofResponse, ProofWithdrawalRequest, ProofWithdrawalValue,
            ProofsWithdrawalResponse, WithdrawalIdQuery,
        },
        state::AppState,
    },
    proof::generate_balance_withdrawal_proof_job,
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use intmax2_zkp::common::witness::withdrawal_witness::WithdrawalWitness;

#[get("/proof/withdrawal/{request_id}")]
async fn get_proof(
    query_params: web::Path<String>,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let request_id = &*query_params;
    let proof = redis::Cmd::get(&withdrawal_token_proof_request_id(request_id))
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

#[get("/proofs/withdrawal")]
async fn get_proofs(
    req: HttpRequest,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder, actix_web::Error> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let query_string = req.query_string();
    let ids_query = serde_qs::from_str::<WithdrawalIdQuery>(query_string);

    let request_ids = match ids_query {
        Ok(query) => query.request_ids,
        Err(e) => {
            log::warn!("Failed to deserialize query: {:?}", e);
            return Ok(HttpResponse::BadRequest().body("Invalid query parameters"));
        }
    };

    let mut proofs: Vec<ProofWithdrawalValue> = Vec::new();
    for request_id in &request_ids {
        let some_proof = redis::Cmd::get(&withdrawal_token_proof_request_id(request_id))
            .query_async::<_, Option<String>>(&mut conn)
            .await
            .map_err(actix_web::error::ErrorInternalServerError)?;
        if let Some(proof) = some_proof {
            proofs.push(ProofWithdrawalValue {
                request_id: (*request_id).to_string(),
                proof,
            });
        }
    }

    let response = ProofsWithdrawalResponse {
        success: true,
        proofs,
        error_message: None,
    };

    Ok(HttpResponse::Ok().json(response))
}

#[post("/proof/withdrawal")]
async fn generate_proof(
    req: web::Json<ProofWithdrawalRequest>,
    redis: web::Data<redis::Client>,
    state: web::Data<AppState>,
) -> Result<impl Responder> {
    let mut redis_conn = redis
        .get_async_connection()
        .await
        .map_err(error::ErrorInternalServerError)?;

    // let request_id = uuid::Uuid::new_v4();
    let request_id = req.request_id.clone();
    let full_request_id = withdrawal_token_proof_request_id(&request_id.to_string());
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

    let balance_circuit_data = state
        .balance_processor
        .get()
        .ok_or_else(|| error::ErrorInternalServerError("balance processor not initialized"))?
        .balance_circuit
        .data
        .verifier_data();
    let single_withdrawal_circuit = state
        .single_withdrawal_circuit
        .get()
        .ok_or_else(|| error::ErrorInternalServerError("balance processor not initialized"))?;

    // Validation check of balance_witness
    let transfer_witness = req.transfer_witness.clone();
    let balance_proof = decode_plonky2_proof(&req.balance_proof, &balance_circuit_data)
        .map_err(error::ErrorInternalServerError)
        .expect("balance proof decoding failed");
    balance_circuit_data
        .verify(balance_proof.clone())
        .map_err(error::ErrorInternalServerError)
        .expect("balance proof verification failed");
    // if transfer_witness.transfers.len() != NUM_TRANSFERS_IN_TX {
    //     println!(
    //         "Invalid number of transfers: {}",
    //         transfer_witness.transfers.len()
    //     );
    //     return Err(error::ErrorBadRequest("Invalid number of transfers"));
    // }
    // if withdrawal_witness.prev_balances.len() != NUM_TRANSFERS_IN_TX {
    //     println!(
    //         "Invalid number of prev_balances: {}",
    //         transfer_witness.prev_balances.len()
    //     );
    //     return Err(error::ErrorBadRequest("Invalid number of prev_balances"));
    // }
    // if withdrawal_witness.asset_merkle_proofs.len() != NUM_TRANSFERS_IN_TX {
    //     println!(
    //         "Invalid number of asset_merkle_proofs: {}",
    //         withdrawal_witness.asset_merkle_proofs.len()
    //     );
    //     return Err(error::ErrorBadRequest(
    //         "Invalid number of asset_merkle_proofs",
    //     ));
    // }

    // let validity_pis =
    //     ValidityPublicInputs::from_pis(&balance_update_witness.validity_proof.public_inputs);
    // if validity_pis != send_witness.tx_witness.validity_pis {
    //     return Err(error::ErrorBadRequest("validity proof pis mismatch"));
    // }

    let withdrawal_witness = WithdrawalWitness {
        transfer_witness,
        balance_proof,
    };
    let response = generate_balance_withdrawal_proof_job(
        &withdrawal_witness,
        state
            .balance_processor
            .get()
            .expect("balance processor not initialized"),
        &single_withdrawal_circuit,
    );
    // let response = generate_balance_spend_proof_job(
    //     &send_witness,
    //     state
    //         .balance_processor
    //         .get()
    //         .expect("balance processor not initialized"),
    // );

    // // Spawn a new task to generate the proof
    // actix_web::rt::spawn(async move {
    //     let response = generate_balance_spend_proof_job(
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

fn withdrawal_token_proof_request_id(request_id: &str) -> String {
    format!("balance-validity/withdrawal/{}", request_id)
}
