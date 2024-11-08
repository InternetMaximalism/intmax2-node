use crate::{
    app::{
        encode::decode_plonky2_proof,
        interface::{
            DepositHashQuery, ProofDepositRequest, ProofDepositValue, ProofResponse,
            ProofsDepositResponse,
        },
        state::AppState,
    },
    proof::{generate_receive_deposit_proof_job, validate_witness},
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use intmax2_zkp::{
    circuits::balance::balance_pis::BalancePublicInputs,
    common::public_state::PublicState,
    ethereum_types::{u256::U256, u32limb_trait::U32LimbTrait},
};

#[get("/proof/{public_key}/deposit/{request_id}")]
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
    let proof = redis::Cmd::get(get_receive_deposit_request_id(
        &public_key.to_hex(),
        request_id,
    ))
    .query_async::<_, Option<String>>(&mut conn)
    .await
    .map_err(error::ErrorInternalServerError)?;

    if proof.is_none() {
        let response = ProofResponse {
            success: false,
            request_id: request_id.clone(),
            proof: None,
            error_message: Some(format!("balance proof is not generated",)),
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

#[get("/proofs/{public_key}/deposit")]
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
    let ids_query = serde_qs::from_str::<DepositHashQuery>(query_string);

    let request_ids: Vec<String> = match ids_query {
        Ok(query) => query.request_ids,
        Err(e) => {
            log::warn!("Failed to deserialize query: {:?}", e);
            return Ok(HttpResponse::BadRequest().body("Invalid query parameters"));
        }
    };

    let mut proofs: Vec<ProofDepositValue> = Vec::new();
    for request_id in &request_ids {
        let some_proof = redis::Cmd::get(&get_receive_deposit_request_id(
            &public_key.to_hex(),
            request_id,
        ))
        .query_async::<_, Option<String>>(&mut conn)
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;
        if let Some(proof) = some_proof {
            proofs.push(ProofDepositValue {
                request_id: (*request_id).to_string(),
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

#[post("/proof/{public_key}/deposit")]
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

    // let request_id = req
    //     .receive_deposit_witness
    //     .deposit_witness
    //     .deposit
    //     .hash()
    //     .to_string();
    let request_id = req.request_id.clone();
    let full_request_id = get_receive_deposit_request_id(&public_key.to_hex(), &request_id);
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
            decode_plonky2_proof(req_prev_balance_proof, &balance_circuit_data)
                .map_err(error::ErrorInternalServerError)?;
        balance_circuit_data
            .verify(prev_balance_proof.clone())
            .map_err(error::ErrorInternalServerError)?;

        Some(prev_balance_proof)
    } else {
        None
    };

    let receive_deposit_witness = req.receive_deposit_witness.clone();
    let public_state = if let Some(prev_balance_proof) = &prev_balance_proof {
        println!("not genesis");
        BalancePublicInputs::from_pis(&prev_balance_proof.public_inputs).public_state
    } else {
        println!("genesis");
        PublicState::genesis()
    };

    validate_witness(
        public_key,
        &public_state,
        &receive_deposit_witness,
        &prev_balance_proof,
    )
    .map_err(error::ErrorInternalServerError)?;

    println!(
        "asset Merkle proof: siblings = {:?}",
        receive_deposit_witness
            .private_transition_witness
            .asset_merkle_proof
    );
    // for (i, sibling) in receive_deposit_witness
    //     .private_witness
    //     .asset_merkle_proof
    //     .0
    //     .siblings
    //     .iter()
    //     .enumerate()
    // {
    //     println!(
    //         "asset Merkle proof: siblings[{}] = {:?}",
    //         i,
    //         sibling.to_string()
    //     );
    // }

    // println!(
    //     "prev asset leaf: {:?}",
    //     receive_deposit_witness.private_witness.prev_asset_leaf
    // );
    // println!(
    //     "prev asset leaf hash: {:?}",
    //     receive_deposit_witness
    //         .private_witness
    //         .prev_asset_leaf
    //         .hash()
    //         .to_string()
    // );
    // println!(
    //     "token index: {}",
    //     receive_deposit_witness.private_witness.token_index
    // );
    // println!(
    //     "asset tree root: {}",
    //     receive_deposit_witness
    //         .private_witness
    //         .prev_private_state
    //         .asset_tree_root
    //         .to_string()
    // );

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
        let response = generate_receive_deposit_proof_job(
            full_request_id,
            public_key,
            prev_balance_proof,
            &receive_deposit_witness,
            state
                .balance_processor
                .get()
                .ok_or_else(|| error::ErrorInternalServerError("balance processor not initialized"))
                .expect("Failed to get balance processor"),
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

fn get_receive_deposit_request_id(public_key: &str, deposit_hash: &str) -> String {
    format!("balance-validity/{}/deposit/{}", public_key, deposit_hash)
}
