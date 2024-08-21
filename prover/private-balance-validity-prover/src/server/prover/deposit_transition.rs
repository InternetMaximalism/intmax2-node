use crate::{
    app::{
        config,
        encode::decode_plonky2_proof,
        interface::{
            DepositIndexQuery, ProofDepositRequest, ProofDepositValue, ProofResponse,
            ProofsDepositResponse,
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

    let deposit_index = query_params
        .1
        .parse::<usize>()
        .map_err(error::ErrorInternalServerError)?;
    let proof = redis::Cmd::get(&get_receive_deposit_request_id(
        &public_key.to_hex(),
        deposit_index,
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
    let ids_query = serde_qs::from_str::<DepositIndexQuery>(query_string);
    let deposit_indices: Vec<String>;

    match ids_query {
        Ok(query) => {
            deposit_indices = query.deposit_indices;
        }
        Err(e) => {
            log::warn!("Failed to deserialize query: {:?}", e);
            return Ok(HttpResponse::BadRequest().body("Invalid query parameters"));
        }
    }

    let mut proofs: Vec<ProofDepositValue> = Vec::new();
    for deposit_index in &deposit_indices {
        let deposit_index_usize = deposit_index.parse::<usize>().unwrap();
        let request_id = get_receive_deposit_request_id(&public_key.to_hex(), deposit_index_usize);
        let some_proof = redis::Cmd::get(&request_id)
            .query_async::<_, Option<String>>(&mut conn)
            .await
            .map_err(actix_web::error::ErrorInternalServerError)?;
        if let Some(proof) = some_proof {
            proofs.push(ProofDepositValue {
                deposit_index: (*deposit_index).to_string(),
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

    let deposit_index = req.receive_deposit_witness.deposit_witness.deposit_index;
    let request_id = get_receive_deposit_request_id(&public_key.to_hex(), deposit_index);
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

    let balance_circuit_data = state
        .balance_processor
        .get()
        .ok_or_else(|| error::ErrorInternalServerError("balance processor not initialized"))?
        .balance_circuit
        .data
        .verifier_data();
    let prev_balance_proof = if let Some(req_prev_balance_proof) = &req.prev_balance_proof {
        log::debug!("requested proof size: {}", req_prev_balance_proof.len());
        let prev_validity_proof =
            decode_plonky2_proof(req_prev_balance_proof, &balance_circuit_data)
                .map_err(error::ErrorInternalServerError)?;
        balance_circuit_data
            .verify(prev_validity_proof.clone())
            .map_err(error::ErrorInternalServerError)?;

        Some(prev_validity_proof)
    } else {
        None
    };

    let receive_deposit_witness = req.receive_deposit_witness.clone();

    // TODO: Validation check of balance_witness

    // Spawn a new task to generate the proof
    actix_web::rt::spawn(async move {
        let balance_public_inputs =
            BalancePublicInputs::from_pis(&prev_balance_proof.unwrap().public_inputs);
        let response = generate_deposit_transition_proof_job(
            &balance_public_inputs,
            &receive_deposit_witness,
            state
                .balance_processor
                .get()
                .ok_or_else(|| error::ErrorInternalServerError("balance processor not initialized"))
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
            "balance proof (deposit_index: {}) is generating",
            deposit_index
        )),
    };

    Ok(HttpResponse::Ok().json(response))
}

fn get_receive_deposit_request_id(public_key: &str, deposit_index: usize) -> String {
    format!("balance-validity/{}/deposit/{}", public_key, deposit_index)
}
