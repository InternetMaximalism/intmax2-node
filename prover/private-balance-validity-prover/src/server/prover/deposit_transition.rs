use crate::{
    app::{
        config,
        encode::decode_plonky2_proof,
        interface::{
            IdsQuery, ProofDepositRequest, ProofDepositValue, ProofResponse, ProofsDepositResponse,
        },
        state::AppState,
    },
    proof::generate_deposit_transition_proof_job,
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use anyhow::Context as _;
use intmax2_zkp::{
    circuits::balance::balance_pis::BalancePublicInputs,
    ethereum_types::{u256::U256, u32limb_trait::U32LimbTrait},
};
use redis::{ExistenceCheck, SetExpiry, SetOptions};

#[get("/proof/{public_key}/transition/deposit/{deposit_index}")]
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
    let proof = redis::Cmd::get(&get_receive_deposit_request_id(
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

#[get("/proofs/{public_key}/transition/deposit")]
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

    let mut proofs: Vec<ProofDepositValue> = Vec::new();
    for private_commitment in &private_commitments {
        let request_id =
            get_receive_deposit_request_id(&public_key.to_hex(), &private_commitment.to_string());
        let some_proof = redis::Cmd::get(&request_id)
            .query_async::<_, Option<String>>(&mut conn)
            .await
            .map_err(actix_web::error::ErrorInternalServerError)?;
        if let Some(proof) = some_proof {
            proofs.push(ProofDepositValue {
                private_commitment: private_commitment.to_string(),
                proof,
            });
        }
    }

    let response = ProofsDepositResponse {
        success: true,
        proofs,
        error_message: None,
    };

    Ok(HttpResponse::Ok().json(response))
}

#[post("/proof/{public_key}/transition/deposit")]
async fn generate_proof(
    query_params: web::Path<String>,
    req: web::Json<ProofDepositRequest>,
    redis: web::Data<redis::Client>,
    state: web::Data<AppState>,
) -> Result<impl Responder> {
    let mut redis_conn = redis
        .get_async_connection()
        .await
        .map_err(error::ErrorInternalServerError)?;

    let public_key = U256::from_hex(&query_params).expect("failed to parse public key");

    let private_commitment = req
        .receive_deposit_witness
        .private_witness
        .prev_private_state
        .commitment();
    let request_id =
        get_receive_deposit_request_id(&public_key.to_hex(), &private_commitment.to_string());
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

    let balance_verifier_data = state
        .balance_verifier_data
        .get()
        .ok_or_else(|| error::ErrorInternalServerError("balance verifier data not initialized"))?;

    let prev_balance_public_inputs = if let Some(req_prev_balance_proof) = &req.prev_balance_proof {
        log::debug!("requested proof size: {}", req_prev_balance_proof.len());
        let prev_balance_proof =
            decode_plonky2_proof(req_prev_balance_proof, &balance_verifier_data)
                .map_err(error::ErrorInternalServerError)?;
        balance_verifier_data
            .verify(prev_balance_proof.clone())
            .map_err(error::ErrorInternalServerError)?;

        BalancePublicInputs::from_pis(&prev_balance_proof.public_inputs)
    } else {
        BalancePublicInputs::new(public_key)
    };

    let receive_deposit_witness = req.receive_deposit_witness.clone();

    // Spawn a new task to generate the proof
    actix_web::rt::spawn(async move {
        let response = generate_deposit_transition_proof_job(
            &prev_balance_public_inputs,
            &receive_deposit_witness,
            &state
                .receive_deposit_circuit
                .get()
                .ok_or_else(|| {
                    error::ErrorInternalServerError("receive deposit circuit not initialized")
                })
                .expect("Failed to get balance processor"),
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

fn get_receive_deposit_request_id(public_key: &str, deposit_index: &str) -> String {
    format!("balance-validity/{}/deposit/{}", public_key, deposit_index)
}
