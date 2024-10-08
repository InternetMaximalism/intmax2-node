use crate::{
    app::{
        interface::{
            ProofResponse, ProofSpendRequest, ProofSpendValue, ProofsSpentResponse, SpentIdQuery,
        },
        state::AppState,
    },
    proof::{generate_balance_spend_proof, RedisResponse},
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use intmax2_zkp::constants::NUM_TRANSFERS_IN_TX;

#[get("/proof/spend/{request_id}")]
async fn get_proof(
    query_params: web::Path<String>,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let request_id = &*query_params;
    let proof_json = redis::Cmd::get(&spent_token_proof_request_id(request_id))
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

#[get("/proofs/spend")]
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
        Ok(query) => query.request_ids,
        Err(e) => {
            log::warn!("Failed to deserialize query: {:?}", e);
            return Ok(HttpResponse::BadRequest().body("Invalid query parameters"));
        }
    };

    let mut proofs: Vec<ProofSpendValue> = Vec::new();
    for request_id in &request_ids {
        let proof_json = redis::Cmd::get(&spent_token_proof_request_id(request_id))
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
                    let response = ProofsSpentResponse {
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
            proofs.push(ProofSpendValue {
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

#[post("/proof/spend")]
async fn generate_proof(
    req: web::Json<ProofSpendRequest>,
    redis: web::Data<redis::Client>,
    state: web::Data<AppState>,
) -> Result<impl Responder> {
    let mut redis_conn = redis
        .get_async_connection()
        .await
        .map_err(error::ErrorInternalServerError)?;

    // let request_id = uuid::Uuid::new_v4();
    let request_id = req.request_id.clone();
    let full_request_id = spent_token_proof_request_id(&request_id.to_string());
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

    // Validation check of balance_witness
    let spent_token_witness = req.send_witness.clone();
    if spent_token_witness.transfers.len() != NUM_TRANSFERS_IN_TX {
        println!(
            "Invalid number of transfers: {}",
            spent_token_witness.transfers.len()
        );
        return Err(error::ErrorBadRequest("Invalid number of transfers"));
    }
    if spent_token_witness.prev_balances.len() != NUM_TRANSFERS_IN_TX {
        println!(
            "Invalid number of prev_balances: {}",
            spent_token_witness.prev_balances.len()
        );
        return Err(error::ErrorBadRequest("Invalid number of prev_balances"));
    }
    if spent_token_witness.asset_merkle_proofs.len() != NUM_TRANSFERS_IN_TX {
        println!(
            "Invalid number of asset_merkle_proofs: {}",
            spent_token_witness.asset_merkle_proofs.len()
        );
        return Err(error::ErrorBadRequest(
            "Invalid number of asset_merkle_proofs",
        ));
    }

    // let validity_pis =
    //     ValidityPublicInputs::from_pis(&balance_update_witness.validity_proof.public_inputs);
    // if validity_pis != send_witness.tx_witness.validity_pis {
    //     return Err(error::ErrorBadRequest("validity proof pis mismatch"));
    // }

    let response = generate_balance_spend_proof(
        &spent_token_witness,
        state
            .balance_processor
            .get()
            .expect("balance processor not initialized"),
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
            let response = ProofResponse {
                success: false,
                request_id: request_id.to_string(),
                proof: None,
                error_message: Some(format!("Failed to generate proof: {e:?}")),
            };

            Ok(HttpResponse::Ok().json(response))
        }
    }
}

fn spent_token_proof_request_id(request_id: &str) -> String {
    format!("balance-validity/spend/{}", request_id)
}
