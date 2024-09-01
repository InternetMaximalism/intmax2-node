use crate::{
    app::{
        encode::decode_plonky2_proof,
        interface::{
            DepositIndexQuery, ProofDepositRequest, ProofDepositValue, ProofResponse,
            ProofsDepositResponse,
        },
        state::AppState,
    },
    proof::generate_receive_deposit_proof_job,
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use intmax2_zkp::{
    circuits::balance::balance_pis::BalancePublicInputs,
    common::{
        deposit::get_pubkey_salt_hash, public_state::PublicState,
        witness::receive_deposit_witness::ReceiveDepositWitness,
    },
    ethereum_types::{bytes32::Bytes32, u256::U256, u32limb_trait::U32LimbTrait},
    utils::leafable::Leafable,
};

#[get("/proof/{public_key}/deposit/{deposit_index}")]
async fn get_proof(
    query_params: web::Path<(String, String)>,
    redis: web::Data<redis::Client>,
    state: web::Data<AppState>,
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

    if proof.is_none() {
        let response = ProofResponse {
            success: false,
            proof: None,
            public_inputs: None,
            error_message: Some(format!(
                "balance proof is not generated (deposit_index: {})",
                deposit_index
            )),
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
    let decompress_proof = decode_plonky2_proof(&proof.clone().unwrap(), &balance_circuit_data)
        .map_err(error::ErrorInternalServerError)?;
    let public_inputs = BalancePublicInputs::from_pis(&decompress_proof.public_inputs);

    let response = ProofResponse {
        success: true,
        proof,
        public_inputs: Some(public_inputs),
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

    let deposit_index = req.receive_deposit_witness.deposit_witness.deposit_index;
    let request_id = get_receive_deposit_request_id(&public_key.to_hex(), deposit_index);
    log::debug!("request ID: {:?}", request_id);
    let old_proof = redis::Cmd::get(&request_id)
        .query_async::<_, Option<String>>(&mut redis_conn)
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;
    if let Some(old_proof) = old_proof {
        let balance_circuit_data = state
            .balance_processor
            .get()
            .ok_or_else(|| error::ErrorInternalServerError("balance processor not initialized"))?
            .balance_circuit
            .data
            .verifier_data();
        let decompress_proof = decode_plonky2_proof(&old_proof, &balance_circuit_data)
            .map_err(error::ErrorInternalServerError)?;
        let public_inputs = BalancePublicInputs::from_pis(&decompress_proof.public_inputs);

        let response = ProofResponse {
            success: true,
            proof: Some(old_proof),
            public_inputs: Some(public_inputs),
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
    // let public_state = if let Some(prev_balance_proof) = &prev_balance_proof {
    //     println!("not genesis");
    //     BalancePublicInputs::from_pis(&prev_balance_proof.public_inputs).public_state
    // } else {
    //     println!("genesis");
    //     PublicState::genesis()
    // };

    // validate_witness(public_key, &public_state, &receive_deposit_witness)
    //     .map_err(error::ErrorInternalServerError)?;

    // Spawn a new task to generate the proof
    actix_web::rt::spawn(async move {
        let response = generate_receive_deposit_proof_job(
            request_id,
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
        public_inputs: None,
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

fn validate_witness(
    pubkey: U256,
    public_state: &PublicState,
    receive_deposit_witness: &ReceiveDepositWitness,
) -> anyhow::Result<()> {
    let deposit_witness = receive_deposit_witness.deposit_witness.clone();
    let private_transition_witness = receive_deposit_witness.private_witness.clone();

    let deposit_salt = receive_deposit_witness.deposit_witness.deposit_salt;
    let deposit_index = receive_deposit_witness.deposit_witness.deposit_index;
    let deposit = &receive_deposit_witness.deposit_witness.deposit;
    let deposit_merkle_proof = &receive_deposit_witness.deposit_witness.deposit_merkle_proof;
    println!("siblings: {:?}\n", deposit_merkle_proof);
    println!("deposit hash: {}\n", deposit.hash().to_hex());
    println!("deposit index: {}\n", deposit_index);
    println!(
        "deposit tree root: {}\n",
        public_state.deposit_tree_root.to_hex()
    );

    let pubkey_salt_hash = get_pubkey_salt_hash(pubkey, deposit_salt);
    if pubkey_salt_hash != deposit.pubkey_salt_hash {
        anyhow::bail!("pubkey_salt_hash not match");
    }

    let result =
        deposit_merkle_proof.verify(&deposit, deposit_index, public_state.deposit_tree_root);
    if !result.is_ok() {
        anyhow::bail!("Invalid deposit merkle proof");
    }

    let deposit = deposit_witness.deposit.clone();
    let nullifier: Bytes32 = deposit.poseidon_hash().into();
    if nullifier != private_transition_witness.nullifier {
        println!("deposit: {:?}", deposit);
        println!("nullifier: {}", nullifier);
        println!(
            "private_transition_witness.nullifier: {}",
            private_transition_witness.nullifier
        );
        anyhow::bail!("nullifier not match");
    }
    // assert_eq!(deposit.token_index, private_transition_witness.token_index);
    if deposit.token_index != private_transition_witness.token_index {
        println!("token_index: {}", deposit.token_index);
        println!(
            "private_transition_witness.token_index: {}",
            private_transition_witness.token_index
        );
        anyhow::bail!("token_index not match");
    }
    // assert_eq!(deposit.amount, private_transition_witness.amount);
    if deposit.amount != private_transition_witness.amount {
        println!("amount: {}", deposit.amount);
        println!(
            "private_transition_witness.amount: {}",
            private_transition_witness.amount
        );
        anyhow::bail!("amount not match");
    }

    Ok(())
}
