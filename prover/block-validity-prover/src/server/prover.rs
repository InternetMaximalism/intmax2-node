use crate::{
    app::{
        encode::decode_plonky2_proof,
        interface::{BlockHashQuery, ProofRequest, ProofResponse, ProofValue, ProofsResponse},
        state::AppState,
    },
    proof::generate_block_validity_proof_job,
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use intmax2_zkp::common::witness::validity_witness::ValidityWitness;

#[get("/proof/{id}")]
async fn get_proof(
    id: web::Path<String>,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let proof = redis::Cmd::get(&get_request_id(&id))
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
        let some_proof: Option<String> = redis::Cmd::get(&request_id)
            .query_async::<_, Option<String>>(&mut conn)
            .await
            .map_err(actix_web::error::ErrorInternalServerError)?;
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
            proof: None,
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

    let validity_witness = ValidityWitness::decompress(&req.validity_witness);
    let request_id = get_request_id(&block_hash);

    // TODO: Validation check of validity_witness

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
                log::error!("Proof generation completed");
                Ok(v)
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
        error_message: Some("validity proof is generating".to_string()),
    };

    Ok(HttpResponse::Ok().json(response))
}

fn get_request_id(block_hash: &str) -> String {
    format!("block-validity/{}", block_hash)
}
