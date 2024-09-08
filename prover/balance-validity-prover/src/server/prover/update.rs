use crate::{
    app::{
        encode::decode_plonky2_proof,
        interface::{
            ProofResponse, ProofUpdateRequest, ProofUpdateValue, ProofsUpdateResponse,
            UpdateIdQuery,
        },
        state::AppState,
    },
    proof::generate_balance_update_proof_job,
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use intmax2_zkp::{
    circuits::{
        balance::balance_pis::BalancePublicInputs, validity::validity_pis::ValidityPublicInputs,
    },
    common::witness::update_witness::UpdateWitness,
    ethereum_types::{u256::U256, u32limb_trait::U32LimbTrait},
};

#[get("/proof/{public_key}/update/{block_hash}")]
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
    let proof = redis::Cmd::get(&get_balance_update_request_id(
        &public_key.to_hex(),
        &request_id,
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

#[get("/proofs/{public_key}/update")]
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
    let ids_query = serde_qs::from_str::<UpdateIdQuery>(query_string);
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

    let mut proofs: Vec<ProofUpdateValue> = Vec::new();
    for block_hash in &block_hashes {
        let some_proof = redis::Cmd::get(&get_balance_update_request_id(
            &public_key.to_hex(),
            block_hash,
        ))
        .query_async::<_, Option<String>>(&mut conn)
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;
        if let Some(proof) = some_proof {
            proofs.push(ProofUpdateValue {
                block_hash: (*block_hash).to_string(),
                proof,
            });
        }
    }

    let response = ProofsUpdateResponse {
        success: true,
        proofs,
        error_message: None,
    };

    Ok(HttpResponse::Ok().json(response))
}

#[post("/proof/{public_key}/update")]
async fn generate_proof(
    query_params: web::Path<String>,
    req: web::Json<ProofUpdateRequest>,
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
    let validity_public_inputs = ValidityPublicInputs::from_pis(&validity_proof.public_inputs);
    let balance_update_witness = UpdateWitness {
        validity_proof,
        block_merkle_proof: req.balance_update_witness.block_merkle_proof.clone(),
        account_membership_proof: req.balance_update_witness.account_membership_proof.clone(),
    };

    let request_id = validity_public_inputs.public_state.block_hash.to_hex();
    let full_request_id = get_balance_update_request_id(&public_key.to_hex(), &request_id);
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

    // TODO: Validation check of balance_witness

    // there is no send tx till the last block
    let prev_balance_pis = if prev_balance_proof.is_some() {
        BalancePublicInputs::from_pis(&prev_balance_proof.as_ref().unwrap().public_inputs)
    } else {
        BalancePublicInputs::new(public_key)
    };
    let last_block_number = balance_update_witness.account_membership_proof.get_value();
    let prev_public_state = &prev_balance_pis.public_state;
    println!("last_block_number: {}", last_block_number);
    println!(
        "balance_update_witness.account_membership_proof.is_included: {}",
        balance_update_witness.account_membership_proof.is_included
    );
    println!(
        "prev_public_state.block_number: {}",
        prev_public_state.block_number
    );
    let encoded_prev_balance_pis = serde_json::to_string(&prev_balance_pis).unwrap();
    println!("encoded_prev_balance_pis: {}", encoded_prev_balance_pis);
    if last_block_number > prev_balance_pis.public_state.block_number as u64 {
        log::warn!(
            "No send tx till the last block: {} > {}",
            last_block_number,
            prev_public_state.block_number
        );
        return Err(error::ErrorInternalServerError(
            "No send tx till the last block",
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
        let response = generate_balance_update_proof_job(
            full_request_id.clone(),
            public_key,
            prev_balance_proof,
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

fn get_balance_update_request_id(public_key: &str, block_hash: &str) -> String {
    format!("balance-validity/{}/update/{}", public_key, block_hash)
}
