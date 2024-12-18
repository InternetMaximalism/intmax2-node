use crate::{
    app::{
        encode::decode_plonky2_proof,
        errors::REDIS_CONNECTION_ERROR,
        interface::{BlockHashQuery, ProofRequest, ProofResponse, ProofValue, ProofsResponse},
        state::AppState,
    },
    proof::generate_block_validity_proof_job,
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use intmax2_zkp::{
    circuits::validity::validity_pis::ValidityPublicInputs,
    common::witness::validity_witness::ValidityWitness,
};

#[get("/proof/{id}")]
async fn get_proof(
    id: web::Path<String>,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(|_| actix_web::error::ErrorInternalServerError(REDIS_CONNECTION_ERROR))?;

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
        .map_err(|_| error::ErrorInternalServerError(REDIS_CONNECTION_ERROR))?;

    let query_string = req.query_string();
    let ids_query: BlockHashQuery = serde_qs::from_str(query_string)
        .map_err(|e| error::ErrorBadRequest(format!("Failed to deserialize query: {e:?}")))?;
    let block_hashes: Vec<String> = ids_query.block_hashes;

    let mut proofs: Vec<ProofValue> = Vec::new();
    for block_hash in &block_hashes {
        let request_id = get_request_id(&block_hash);
        let some_proof: Option<String> = redis::Cmd::get(&request_id)
            .query_async::<_, Option<String>>(&mut conn)
            .await
            .map_err(error::ErrorInternalServerError)?;
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
        .map_err(|_| error::ErrorInternalServerError(REDIS_CONNECTION_ERROR))?;

    if req.block_hash.is_empty() {
        return Err(error::ErrorBadRequest(anyhow::anyhow!(
            "block_hash is empty"
        )));
    }

    let old_proof = redis::Cmd::get(&get_request_id(&req.block_hash))
        .query_async::<_, Option<String>>(&mut redis_conn)
        .await
        .map_err(error::ErrorInternalServerError)?;
    if old_proof.is_some() {
        let response = ProofResponse {
            success: true,
            proof: old_proof.clone(),
            error_message: Some("validity proof already exists".to_string()),
        };

        return Ok(HttpResponse::Ok().json(response));
    }

    let block_hash = req.block_hash.to_lowercase();
    let s = block_hash.strip_prefix("0x").unwrap_or(&block_hash);
    let ok = s.chars().all(|c| c.is_digit(16));
    if !ok {
        return Err(error::ErrorBadRequest(format!(
            "Invalid block hash: {block_hash}"
        )));
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
                .map_err(error::ErrorBadRequest)?;
        validity_circuit_data
            .verify(prev_validity_proof.clone())
            .map_err(error::ErrorBadRequest)?;

        Some(prev_validity_proof)
    } else {
        None
    };

    let validity_witness = if let Some(req_plain_validity_witness) = &req.plain_validity_witness {
        req_plain_validity_witness.clone()
    } else {
        ValidityWitness::decompress(&req.validity_witness.clone().unwrap())
    };

    // let encoded_validity_witness =
    //     serde_json::to_string(&validity_witness).map_err(error::ErrorInternalServerError)?;
    // let mut file = std::fs::File::create("encoded_validity_witness.json").unwrap();
    // file.write_all(encoded_validity_witness.as_bytes()).unwrap();
    // let encoded_compressed_validity_witness =
    //     serde_json::to_string(&req.validity_witness).map_err(error::ErrorInternalServerError)?;
    // let mut file = std::fs::File::create("encoded_compressed_validity_witness.json").unwrap();
    // file.write_all(encoded_compressed_validity_witness.as_bytes())
    //     .unwrap();

    let new_pis = validity_witness.to_validity_pis().unwrap();
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
        return Err(error::ErrorBadRequest(format!(
            "account tree root is mismatch: {} != {}",
            prev_pis.public_state.account_tree_root,
            validity_witness.block_witness.prev_account_tree_root
        )));
    }
    if prev_pis.public_state.block_tree_root != validity_witness.block_witness.prev_block_tree_root
    {
        return Err(error::ErrorBadRequest(format!(
            "block tree root is mismatch: {} != {}",
            prev_pis.public_state.block_tree_root,
            validity_witness.block_witness.prev_block_tree_root
        )));
    }

    // Spawn a new task to generate the proof
    actix_web::rt::spawn(async move {
        let response = generate_block_validity_proof_job(
            request_id,
            prev_validity_proof,
            validity_witness,
            state
                .validity_processor
                .get()
                .expect("validity processor not initialized"),
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
