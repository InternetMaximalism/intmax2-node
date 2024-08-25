use crate::{
    app::{
        encode::decode_plonky2_proof,
        interface::{
            FraudIdQuery, FraudProofRequest, GenerateProofResponse, ProofResponse, ProofValue,
            ProofsResponse,
        },
        state::AppState,
    },
    proof::generate_fraud_proof_job,
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};
use intmax2_zkp::{
    circuits::validity::validity_pis::ValidityPublicInputs,
    ethereum_types::{address::Address, u32limb_trait::U32LimbTrait as _},
};

#[get("/proof/fraud/{id}")]
async fn get_proof(
    id: web::Path<String>,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let request_id = get_fraud_request_id(&id);
    let proof = redis::Cmd::get(&request_id)
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

#[get("/proofs/fraud")]
async fn get_proofs(req: HttpRequest, redis: web::Data<redis::Client>) -> Result<impl Responder> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let query_string = req.query_string();
    let ids_query: Result<FraudIdQuery, _> = serde_qs::from_str(query_string);
    let ids: Vec<String>;

    match ids_query {
        Ok(query) => {
            ids = query.ids;
        }
        Err(e) => {
            log::warn!("Failed to deserialize query: {:?}", e);
            return Ok(HttpResponse::BadRequest().body("Invalid query parameters"));
        }
    }

    let mut proofs: Vec<ProofValue> = Vec::new();
    for id in &ids {
        let request_id = get_fraud_request_id(&id);
        let some_proof = redis::Cmd::get(&request_id)
            .query_async::<_, Option<String>>(&mut conn)
            .await
            .map_err(actix_web::error::ErrorInternalServerError)?;
        if let Some(proof) = some_proof {
            proofs.push(ProofValue {
                id: (*id).to_string(),
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

#[post("/proof/fraud")]
async fn generate_proof(
    req: web::Json<FraudProofRequest>,
    redis: web::Data<redis::Client>,
    state: web::Data<AppState>,
) -> Result<impl Responder> {
    let mut redis_conn = redis
        .get_async_connection()
        .await
        .map_err(error::ErrorInternalServerError)?;

    let validity_circuit_data = state
        .validity_circuit
        .get()
        .ok_or_else(|| error::ErrorInternalServerError("validity circuit is not initialized"))?
        .data
        .verifier_data();
    let validity_proof = decode_plonky2_proof(&req.validity_proof, &validity_circuit_data)
        .map_err(error::ErrorInternalServerError)?;

    let challenger = Address::from_hex(&req.challenger).map_err(error::ErrorInternalServerError)?;

    let block_hash = ValidityPublicInputs::from_pis(&validity_proof.public_inputs)
        .public_state
        .block_hash;
    let request_id = get_fraud_request_id(&block_hash.to_string());
    let old_proof = redis::Cmd::get(&request_id)
        .query_async::<_, Option<String>>(&mut redis_conn)
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;
    if old_proof.is_some() {
        let response = ProofResponse {
            success: true,
            proof: None,
            error_message: Some("Fraud proof already exists".to_string()),
        };

        return Ok(HttpResponse::Ok().json(response));
    }

    // Spawn a new task to generate the proof
    actix_web::rt::spawn(async move {
        let response = generate_fraud_proof_job(
            request_id,
            challenger,
            &validity_proof,
            state
                .fraud_processor
                .get()
                .expect("fraud circuit is not initialized"),
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

    Ok(HttpResponse::Ok().json(GenerateProofResponse {
        success: true,
        message: format!(
            "fraud proof (block hash: {}) is generating",
            block_hash.to_string()
        ),
    }))
}

fn get_fraud_request_id(block_hash: &str) -> String {
    format!("fraud/{}", block_hash)
}
