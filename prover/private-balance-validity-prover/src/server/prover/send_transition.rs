use crate::{
    app::{
        config,
        interface::{
            IdsQuery, ProofResponse, ProofSendRequest, ProofSendValue, ProofsSendResponse,
        },
        state::AppState,
    },
    proof::generate_spent_proof_job,
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use anyhow::Context as _;
use intmax2_zkp::ethereum_types::{u256::U256, u32limb_trait::U32LimbTrait};
use redis::{ExistenceCheck, SetExpiry, SetOptions};

#[get("/proof/{public_key}/transition/send/{block_hash}")]
async fn get_proof(
    query_params: web::Path<(String, String)>,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let public_key = U256::from_hex(&query_params.0).expect("failed to parse public key");
    let private_commitment = &query_params.1;
    let proof = redis::Cmd::get(&get_balance_send_request_id(
        &public_key.to_hex(),
        private_commitment,
    ))
    .query_async::<_, Option<String>>(&mut conn)
    .await
    .map_err(error::ErrorInternalServerError)?;

    let response = ProofResponse {
        success: true,
        proof,
        error_message: None,
    };

    Ok(HttpResponse::Ok().json(response))
}

#[get("/proofs/{public_key}/transition/send")]
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
    let private_commitments_query = serde_qs::from_str::<IdsQuery>(query_string);
    let private_commitments: Vec<String>;

    match private_commitments_query {
        Ok(query) => {
            private_commitments = query.ids;
        }
        Err(e) => {
            log::warn!("Failed to deserialize query: {:?}", e);
            return Ok(HttpResponse::BadRequest().body("Invalid query parameters"));
        }
    }

    let mut proofs: Vec<ProofSendValue> = Vec::new();
    for private_commitment in &private_commitments {
        let request_id =
            get_balance_send_request_id(&public_key.to_hex(), &private_commitment.to_string());
        let some_proof = redis::Cmd::get(&request_id)
            .query_async::<_, Option<String>>(&mut conn)
            .await
            .map_err(actix_web::error::ErrorInternalServerError)?;
        if let Some(proof) = some_proof {
            proofs.push(ProofSendValue {
                private_commitment: private_commitment.to_string(),
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

#[post("/proof/{public_key}/transition/send")]
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

    let private_commitment = req.send_witness.prev_private_state.commitment();
    let request_id =
        get_balance_send_request_id(&public_key.to_hex(), &private_commitment.to_string());
    log::debug!("request ID: {:?}", request_id);
    let old_proof = redis::Cmd::get(&request_id)
        .query_async::<_, Option<String>>(&mut redis_conn)
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;
    if old_proof.is_some() {
        let response = ProofResponse {
            success: true,
            proof: None,
            error_message: Some("balance proof already requested".to_string()),
        };

        return Ok(HttpResponse::Ok().json(response));
    }

    // Spawn a new task to generate the proof
    actix_web::rt::spawn(async move {
        let response = generate_spent_proof_job(
            &req.send_witness,
            state
                .spent_circuit
                .get()
                .expect("spent circuit not initialized"),
        );

        match response {
            Ok(proof) => {
                let opts = SetOptions::default()
                    .conditional_set(ExistenceCheck::NX)
                    .get(true)
                    .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

                let _ = redis::Cmd::set_options(&request_id, proof, opts)
                    .query_async::<_, Option<String>>(&mut redis_conn)
                    .await
                    .with_context(|| "Failed to set proof")?;

                log::info!("Proof generation completed");
                Ok(())
            }
            Err(e) => {
                log::error!("Failed to generate proof: {:?}", e);
                Err(e)
            }
        }
    });

    let response = ProofResponse {
        success: true,
        proof: None,
        error_message: Some(format!(
            "balance proof (id: {}) is generating",
            private_commitment
        )),
    };

    Ok(HttpResponse::Ok().json(response))
}

fn get_balance_send_request_id(public_key: &str, block_hash: &str) -> String {
    format!("balance-validity/{}/send/{}", public_key, block_hash)
}
