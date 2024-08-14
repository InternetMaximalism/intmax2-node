use crate::{
    app::{
        interface::{
            ProofResponse, ProofTransferRequest, ProofTransferValue,
            ProofsTransferResponse, TransferIdQuery,
        },
        encode::decode_plonky2_proof,
        state::AppState,
    },
    proof::generate_balance_transfer_proof_job,
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use intmax2_zkp::{
    circuits::balance::balance_pis::BalancePublicInputs,
    ethereum_types::{u256::U256, u32limb_trait::U32LimbTrait},
};

#[get("/proof/{public_key}/transfer/{block_hash}")]
async fn get_proof(
    query_params: web::Path<(String, String)>,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let public_key = U256::from_hex(&query_params.0).expect("failed to parse public key");

    let block_hash = &query_params.1;
    let proof = redis::Cmd::get(&get_balance_transfer_request_id(
        &public_key.to_hex(),
        block_hash,
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

#[get("/proofs/{public_key}/transfer")]
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
    let block_hashes: Vec<String>;

    match ids_query {
        Ok(query) => {
            block_hashes = query.block_hashes;
        }
        Err(e) => {
            eprintln!("Failed to deserialize query: {:?}", e);
            return Ok(HttpResponse::BadRequest().body("Invalid query parameters"));
        }
    }

    let mut proofs: Vec<ProofTransferValue> = Vec::new();
    for block_hash in &block_hashes {
        let request_id = get_balance_transfer_request_id(&public_key.to_hex(), block_hash);
        let some_proof = redis::Cmd::get(&request_id)
            .query_async::<_, Option<String>>(&mut conn)
            .await
            .map_err(actix_web::error::ErrorInternalServerError)?;
        if let Some(proof) = some_proof {
            proofs.push(ProofTransferValue {
                block_hash: (*block_hash).to_string(),
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

#[post("/proof/{public_key}/transfer")]
async fn generate_proof(
    query_params: web::Path<String>,
    req: web::Json<ProofTransferRequest>,
    redis: web::Data<redis::Client>,
    state: web::Data<AppState>,
) -> Result<impl Responder> {
    println!("POST /proof");
    let mut redis_conn = redis
        .get_async_connection()
        .await
        .map_err(error::ErrorInternalServerError)?;

    let public_key = U256::from_hex(&query_params).expect("failed to parse public key");

    let balance_circuit_data = state
        .balance_processor
        .get()
        .ok_or_else(|| error::ErrorInternalServerError("validity circuit not initialized"))?
        .balance_circuit
        .data
        .verifier_data();

    let receive_transfer_witness = req
        .receive_transfer_witness
        .decode(&balance_circuit_data)
        .map_err(error::ErrorInternalServerError)?;
    balance_circuit_data
        .verify(receive_transfer_witness.balance_proof.clone())
        .map_err(error::ErrorInternalServerError)?;
    let balance_public_inputs =
        BalancePublicInputs::from_pis(&receive_transfer_witness.balance_proof.public_inputs);

    let block_hash = balance_public_inputs.public_state.block_hash;
    let request_id = get_balance_transfer_request_id(&public_key.to_hex(), &block_hash.to_hex());
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

    let prev_balance_proof = if let Some(req_prev_balance_proof) = &req.prev_balance_proof {
        println!("requested proof size: {}", req_prev_balance_proof.len());
        let prev_balance_proof =
            decode_plonky2_proof(req_prev_balance_proof, &balance_circuit_data)
                .map_err(error::ErrorInternalServerError)?;
        balance_circuit_data
            .verify(prev_balance_proof.clone())
            .map_err(error::ErrorInternalServerError)?;

        Some(prev_balance_proof)
    } else {
        None
    };

    // TODO: Validation check of balance_witness

    // Spawn a new task to generate the proof
    actix_web::rt::spawn(async move {
        let response = generate_balance_transfer_proof_job(
            request_id,
            public_key,
            prev_balance_proof,
            &receive_transfer_witness,
            state
                .balance_processor
                .get()
                .expect("balance processor not initialized"),
            &mut redis_conn,
        )
        .await;

        match response {
            Ok(v) => {
                println!("Proof generation completed");
                Ok(v)
            }
            Err(e) => {
                eprintln!("Failed to generate proof: {:?}", e);
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

fn get_balance_transfer_request_id(public_key: &str, block_hash: &str) -> String {
    format!("balance-validity/{}/transfer/{}", public_key, block_hash)
}
