use crate::{
    app::{
        config,
        interface::{
            ProofResponse, ProofTransferRequest, ProofTransferValue, ProofsTransferResponse,
            TransferIdQuery,
        },
        state::AppState,
    },
    proof::generate_transfer_transition_proof_job,
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use anyhow::Context as _;
use intmax2_zkp::{
    circuits::balance::balance_pis::BalancePublicInputs,
    ethereum_types::{u256::U256, u32limb_trait::U32LimbTrait},
};
use redis::{ExistenceCheck, SetExpiry, SetOptions};

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

    let private_commitment = &query_params.1;
    let proof = redis::Cmd::get(&get_balance_transfer_request_id(
        &public_key.to_hex(),
        &private_commitment,
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

#[get("/proofs/{public_key}/transition/transfer")]
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
        let request_id = get_balance_transfer_request_id(&public_key.to_hex(), private_commitment);
        let some_proof = redis::Cmd::get(&request_id)
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

#[post("/proof/{public_key}/transition/transfer")]
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

    let balance_circuit_verifier_data = state.balance_verifier_data.get().ok_or_else(|| {
        error::ErrorInternalServerError("verifier data of balance circuit not initialized")
    })?;

    let receive_transfer_witness = req
        .receive_transfer_witness
        .decode(&balance_circuit_verifier_data)
        .map_err(error::ErrorInternalServerError)?;
    let balance_public_inputs =
        BalancePublicInputs::from_pis(&receive_transfer_witness.balance_proof.public_inputs);
    let private_commitment = balance_public_inputs.private_commitment;
    let request_id =
        get_balance_transfer_request_id(&public_key.to_hex(), &private_commitment.to_string());
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

    let prev_balance_public_inputs = serde_json::from_str(&req.prev_balance_public_inputs)
        .map_err(actix_web::error::ErrorInternalServerError)?;

    // Spawn a new task to generate the proof
    actix_web::rt::spawn(async move {
        let response = generate_transfer_transition_proof_job(
            &prev_balance_public_inputs,
            &receive_transfer_witness,
            state
                .balance_transition_processor
                .get()
                .expect("balance transition processor not initialized"),
            &state
                .balance_verifier_data
                .get()
                .expect("verifier data of balance circuit not initialized"),
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
            "balance proof (private_commitment: {}) is generating",
            private_commitment
        )),
    };

    Ok(HttpResponse::Ok().json(response))
}

fn get_balance_transfer_request_id(public_key: &str, private_commitment: &str) -> String {
    format!(
        "balance-validity/{}/transfer/{}",
        public_key, private_commitment
    )
}
