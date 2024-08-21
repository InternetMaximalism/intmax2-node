use crate::{
    app::{
        config,
        encode::decode_plonky2_proof,
        interface::{
            ProofResponse, ProofSendRequest, ProofSendValue, ProofsSendResponse, SendIdQuery,
        },
        state::AppState,
    },
    proof::generate_send_transition_proof_job,
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use anyhow::Context as _;
use intmax2_zkp::{
    common::witness::update_witness::UpdateWitness,
    ethereum_types::{u256::U256, u32limb_trait::U32LimbTrait},
};
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

    let proof = redis::Cmd::get(&get_balance_send_request_id(
        &public_key.to_hex(),
        &query_params.1,
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
    let ids_query = serde_qs::from_str::<SendIdQuery>(query_string);
    let block_hashes: Vec<String>;

    match ids_query {
        Ok(query) => {
            block_hashes = query.block_hashes;
        }
        Err(e) => {
            log::warn!("Failed to deserialize query: {:?}", e);
            return Ok(HttpResponse::BadRequest().body("Invalid query parameters"));
        }
    }

    let mut proofs: Vec<ProofSendValue> = Vec::new();
    for block_hash in &block_hashes {
        let request_id = get_balance_send_request_id(&public_key.to_hex(), block_hash);
        let some_proof = redis::Cmd::get(&request_id)
            .query_async::<_, Option<String>>(&mut conn)
            .await
            .map_err(actix_web::error::ErrorInternalServerError)?;
        if let Some(proof) = some_proof {
            proofs.push(ProofSendValue {
                block_hash: (*block_hash).to_string(),
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

    let validity_verifier_data = state
        .validity_verifier_data
        .get()
        .ok_or_else(|| error::ErrorInternalServerError("validity circuit not initialized"))?;
    let encoded_validity_proof = req.balance_update_witness.validity_proof.clone();
    let validity_proof = decode_plonky2_proof(&encoded_validity_proof, &validity_verifier_data)
        .map_err(error::ErrorInternalServerError)?;
    validity_verifier_data
        .verify(validity_proof.clone())
        .map_err(error::ErrorInternalServerError)?;
    let balance_update_witness = UpdateWitness {
        validity_proof,
        block_merkle_proof: req.balance_update_witness.block_merkle_proof.clone(),
        account_membership_proof: req.balance_update_witness.account_membership_proof.clone(),
    };

    // let block_hash = validity_public_inputs.public_state.block_hash;
    let block_hash = U256::from(1); // TODO
    let request_id = get_balance_send_request_id(&public_key.to_hex(), &block_hash.to_hex());
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
        let response = generate_send_transition_proof_job(
            &req.send_witness,
            &balance_update_witness,
            state
                .balance_transition_processor
                .get()
                .expect("balance transition processor not initialized"),
            &state
                .balance_verifier_data
                .get()
                .expect("verifier data of balance circuit not initialized")
                .verifier_only,
            state
                .validity_verifier_data
                .get()
                .expect("validity circuit not initialized"),
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
            "balance proof (block_hash: {}) is generating",
            block_hash
        )),
    };

    Ok(HttpResponse::Ok().json(response))
}

fn get_balance_send_request_id(public_key: &str, block_hash: &str) -> String {
    format!("balance-validity/{}/send/{}", public_key, block_hash)
}
