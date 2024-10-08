use std::io::Write;

use crate::{
    app::{
        encode::decode_plonky2_proof,
        interface::{BlockHashQuery, ProofRequest, ProofResponse, ProofValue, ProofsResponse},
        state::AppState,
    },
    proof::{generate_block_validity_proof_job, RedisResponse},
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use intmax2_zkp::{
    circuits::validity::validity_pis::ValidityPublicInputs,
    common::witness::validity_witness::ValidityWitness,
};

#[get("/proof/{request_id}")]
async fn get_proof(
    request_id: web::Path<String>,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let proof_json = redis::Cmd::get(&get_request_id(&request_id))
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
                error_message: Some("validity proof is not generated".to_string()),
            };

            Ok(HttpResponse::Ok().json(response))
        }
    }
}

#[get("/proofs")]
async fn get_proofs(
    req: HttpRequest,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder, actix_web::Error> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let query_string = req.query_string();
    let ids_query: Result<BlockHashQuery, _> = serde_qs::from_str(query_string);
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

    let mut proofs: Vec<ProofValue> = Vec::new();
    for block_hash in &block_hashes {
        let request_id = get_request_id(&block_hash);
        let proof_json: Option<String> = redis::Cmd::get(&request_id)
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
                    let response = ProofsResponse {
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
            proofs.push(ProofValue {
                block_hash: (*block_hash).to_string(),
                proof,
            });
        }
    }

    let response = ProofsResponse {
        success: true,
        proofs,
        error_message: None,
    };

    Ok(HttpResponse::Ok().json(response))
}

#[post("/proof")]
async fn generate_proof(
    req: web::Json<ProofRequest>,
    redis: web::Data<redis::Client>,
    state: web::Data<AppState>,
) -> Result<impl Responder> {
    let mut redis_conn = redis
        .get_async_connection()
        .await
        .map_err(error::ErrorInternalServerError)?;

    let old_proof = redis::Cmd::get(&get_request_id(&req.block_hash))
        .query_async::<_, Option<String>>(&mut redis_conn)
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;
    if old_proof.is_some() {
        let response = ProofResponse {
            success: true,
            request_id: req.block_hash.clone(),
            proof: old_proof.clone(),
            error_message: Some("validity proof already exists".to_string()),
        };

        return Ok(HttpResponse::Ok().json(response));
    }

    let block_hash = req.block_hash.to_lowercase();
    let s = block_hash.strip_prefix("0x").unwrap_or(&block_hash);
    let ok = s.chars().all(|c| c.is_digit(16));
    if !ok {
        log::warn!("Invalid block hash: {block_hash}");
        return Err(error::ErrorInternalServerError("Invalid block hash"));
    }
    log::debug!("block_hash: {:?}", block_hash);

    let validity_circuit_data = state
        .validity_processor
        .get()
        .ok_or_else(|| error::ErrorInternalServerError("validity processor not initialized"))?
        .validity_circuit
        .data
        .verifier_data();

    let prev_validity_proof = if let Some(req_prev_validity_proof) = &req.prev_validity_proof {
        log::debug!("requested proof size: {}", req_prev_validity_proof.len());
        let prev_validity_proof =
            decode_plonky2_proof(req_prev_validity_proof, &validity_circuit_data)
                .map_err(error::ErrorInternalServerError)?;
        validity_circuit_data
            .verify(prev_validity_proof.clone())
            .map_err(error::ErrorInternalServerError)?;

        Some(prev_validity_proof)
    } else {
        None
    };

    let validity_witness = if let Some(req_plain_validity_witness) = &req.plain_validity_witness {
        req_plain_validity_witness.clone()
    } else {
        ValidityWitness::decompress(&req.validity_witness.clone().unwrap())
    };

    let new_pis = validity_witness.to_validity_pis();
    println!(
        "new_pis block_number: {}",
        new_pis.public_state.block_number
    );
    println!(
        "new_pis prev_account_tree_root: {}",
        new_pis.public_state.prev_account_tree_root
    );
    println!(
        "new_pis account_tree_root: {}",
        new_pis.public_state.account_tree_root
    );
    println!("new_pis is_valid: {}", new_pis.is_valid_block);

    let request_id = get_request_id(&block_hash);

    // TODO: Validation check of validity_witness
    let prev_pis = if prev_validity_proof.is_some() {
        ValidityPublicInputs::from_pis(&prev_validity_proof.as_ref().unwrap().public_inputs)
    } else {
        ValidityPublicInputs::genesis()
    };
    if prev_pis.public_state.account_tree_root
        != validity_witness.block_witness.prev_account_tree_root
    {
        let response = ProofResponse {
            success: false,
            request_id: request_id.clone(),
            proof: None,
            error_message: Some("account tree root is mismatch".to_string()),
        };
        println!(
            "block tree root is mismatch: {} != {}",
            prev_pis.public_state.account_tree_root,
            validity_witness.block_witness.prev_account_tree_root
        );
        return Ok(HttpResponse::Ok().json(response));
    }
    if prev_pis.public_state.block_tree_root != validity_witness.block_witness.prev_block_tree_root
    {
        let response = ProofResponse {
            success: false,
            request_id: request_id.clone(),
            proof: None,
            error_message: Some("block tree root is mismatch".to_string()),
        };
        println!(
            "block tree root is mismatch: {} != {}",
            prev_pis.public_state.block_tree_root,
            validity_witness.block_witness.prev_block_tree_root
        );
        return Ok(HttpResponse::Ok().json(response));
    }

    let response = ProofResponse {
        success: true,
        request_id: request_id.clone(),
        proof: None,
        error_message: Some("validity proof is generating".to_string()),
    };

    // Spawn a new task to generate the proof
    actix_web::rt::spawn(async move {
        let response = generate_block_validity_proof_job(
            request_id,
            prev_validity_proof,
            validity_witness,
            state
                .validity_processor
                .get()
                .ok_or_else(|| {
                    error::ErrorInternalServerError("validity processor not initialized")
                })
                .expect("Failed to get validity processor"),
            &mut redis_conn,
        )
        .await;

        match response {
            Ok(v) => {
                log::info!("Proof generation completed");
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

fn get_request_id(block_hash: &str) -> String {
    format!("block-validity/{}", block_hash)
}
