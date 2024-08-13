use crate::{
    app::{
        interface::{
            ProofResponse, ProofUpdateRequest, ProofUpdateValue, ProofsUpdateResponse,
            UpdateIdQuery,
        },
        state::AppState,
    },
    proof::generate_balance_update_proof_job,
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use base64::prelude::*;
use intmax2_zkp::{
    circuits::validity::validity_pis::ValidityPublicInputs,
    common::witness::update_witness::UpdateWitness,
    ethereum_types::{u256::U256, u32limb_trait::U32LimbTrait},
};
use plonky2::{
    field::goldilocks_field::GoldilocksField,
    plonk::{config::PoseidonGoldilocksConfig, proof::CompressedProofWithPublicInputs},
};

type C = PoseidonGoldilocksConfig;
const D: usize = 2;
type F = GoldilocksField;

#[get("/proof/{public_key}/update/{deposit_index}")]
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
    let proof = redis::Cmd::get(&get_balance_update_request_id(
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
    let ids: Vec<String>;

    match ids_query {
        Ok(query) => {
            ids = query.block_hashes;
        }
        Err(e) => {
            eprintln!("Failed to deserialize query: {:?}", e);
            return Ok(HttpResponse::BadRequest().body("Invalid query parameters"));
        }
    }

    let mut proofs: Vec<ProofUpdateValue> = Vec::new();
    for block_hash in &ids {
        let request_id = get_balance_update_request_id(&public_key.to_hex(), block_hash);
        let some_proof = redis::Cmd::get(&request_id)
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
    println!("POST /proof");
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
    let decoded_validity_proof = BASE64_STANDARD
        .decode(&encoded_validity_proof)
        .map_err(error::ErrorInternalServerError)?;
    let compressed_validity_proof = CompressedProofWithPublicInputs::<F, C, D>::from_bytes(
        decoded_validity_proof,
        &validity_circuit_data.common,
    )
    .map_err(error::ErrorInternalServerError)?;
    let validity_proof = compressed_validity_proof
        .decompress(
            &validity_circuit_data.verifier_only.circuit_digest,
            &validity_circuit_data.common,
        )
        .map_err(error::ErrorInternalServerError)?;
    validity_circuit_data
        .verify(validity_proof.clone())
        .map_err(error::ErrorInternalServerError)?;
    let validity_public_inputs = ValidityPublicInputs::from_pis(&validity_proof.public_inputs);

    let block_hash = validity_public_inputs.public_state.block_hash;
    let request_id = get_balance_update_request_id(&public_key.to_hex(), &block_hash.to_hex());
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
        println!("requested proof size: {}", req_prev_balance_proof.len());
        let decoded_prev_validity_proof = BASE64_STANDARD
            .decode(&req_prev_balance_proof)
            .map_err(error::ErrorInternalServerError)?;
        println!("balance proof size: {}", decoded_prev_validity_proof.len());

        let compressed_prev_validity_proof =
            CompressedProofWithPublicInputs::<F, C, D>::from_bytes(
                decoded_prev_validity_proof,
                &balance_circuit_data.common,
            )
            .map_err(error::ErrorInternalServerError)?;
        let prev_validity_proof = compressed_prev_validity_proof
            .decompress(
                &balance_circuit_data.verifier_only.circuit_digest,
                &balance_circuit_data.common,
            )
            .map_err(error::ErrorInternalServerError)?;
        balance_circuit_data
            .verify(prev_validity_proof.clone())
            .map_err(error::ErrorInternalServerError)?;

        Some(prev_validity_proof)
    } else {
        None
    };

    let encoded_validity_proof = BASE64_STANDARD
        .decode(&req.balance_update_witness.validity_proof)
        .map_err(error::ErrorInternalServerError)?;
    let compressed_validity_proof = CompressedProofWithPublicInputs::from_bytes(
        encoded_validity_proof,
        &validity_circuit_data.common,
    )
    .map_err(error::ErrorInternalServerError)?;
    let validity_proof = compressed_validity_proof
        .decompress(
            &validity_circuit_data.verifier_only.circuit_digest,
            &validity_circuit_data.common,
        )
        .map_err(error::ErrorInternalServerError)?;
    let balance_update_witness = UpdateWitness {
        validity_proof,
        block_merkle_proof: req.balance_update_witness.block_merkle_proof.clone(),
        account_membership_proof: req.balance_update_witness.account_membership_proof.clone(),
    };

    // TODO: Validation check of balance_witness

    // Spawn a new task to generate the proof
    actix_web::rt::spawn(async move {
        let response = generate_balance_update_proof_job(
            request_id,
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

fn get_balance_update_request_id(public_key: &str, block_hash: &str) -> String {
    format!("balance-validity/{}/update/{}", public_key, block_hash)
}
