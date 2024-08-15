use crate::{
    app::{
        encode::decode_plonky2_proof,
        interface::{GenerateProofResponse, WithdrawalIdQuery, WithdrawalProofRequest, ProofResponse, ProofValue, ProofsResponse},
        state::AppState,
    },
    proof::generate_withdrawal_proof_job,
};
use actix_web::{error, get, post, web, HttpRequest, HttpResponse, Responder, Result};

#[get("/proof/withdrawal/{id}")]
async fn get_proof(
    id: web::Path<String>,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let request_id = get_withdrawal_request_id(&id);
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

#[get("/proofs/withdrawal")]
async fn get_proofs(
    req: HttpRequest,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let query_string = req.query_string();
    let ids_query: Result<WithdrawalIdQuery, _> = serde_qs::from_str(query_string);
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
        let request_id = get_withdrawal_request_id(&id);
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

#[post("/proof/withdrawal")]
async fn generate_proof(
    req: web::Json<WithdrawalProofRequest>,
    redis: web::Data<redis::Client>,
    state: web::Data<AppState>,
) -> Result<impl Responder> {
    let mut redis_conn = redis
        .get_async_connection()
        .await
        .map_err(error::ErrorInternalServerError)?;

    let request_id = get_withdrawal_request_id(&req.id);
    let old_proof = redis::Cmd::get(&request_id)
        .query_async::<_, Option<String>>(&mut redis_conn)
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;
    if old_proof.is_some() {
        let response = ProofResponse {
            success: true,
            proof: None,
            error_message: Some("withdrawal proof already exists".to_string()),
        };

        return Ok(HttpResponse::Ok().json(response));
    }

    let withdrawal_circuit_data = state
        .withdrawal_processor
        .get()
        .ok_or_else(|| error::ErrorInternalServerError("withdrawal circuit is not initialized"))?
        .withdrawal_circuit
        .data
        .verifier_data();
    let prev_withdrawal_proof = if let Some(req_prev_withdrawal_proof) = &req.prev_withdrawal_proof
    {
        log::debug!("requested proof size: {}", req_prev_withdrawal_proof.len());
        let prev_withdrawal_proof =
            decode_plonky2_proof(req_prev_withdrawal_proof, &withdrawal_circuit_data)
                .map_err(error::ErrorInternalServerError)?;
        withdrawal_circuit_data
            .verify(prev_withdrawal_proof.clone())
            .map_err(error::ErrorInternalServerError)?;

        Some(prev_withdrawal_proof)
    } else {
        None
    };

    let balance_circuit_data = state
        .balance_circuit
        .get()
        .ok_or_else(|| error::ErrorInternalServerError("withdrawal circuit is not initialized"))?
        .data
        .verifier_data();
    let withdrawal_witness = req
        .withdrawal_witness
        .decode(&balance_circuit_data)
        .map_err(error::ErrorInternalServerError)?;

    // TODO: Validation check of withdrawal_witness

    // Spawn a new task to generate the proof
    actix_web::rt::spawn(async move {
        let response = generate_withdrawal_proof_job(
            request_id,
            prev_withdrawal_proof,
            &withdrawal_witness,
            state
                .withdrawal_processor
                .get()
                .expect("withdrawal circuit is not initialized"),
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
        message: "withdrawal proof is generating".to_string(),
    }))
}

fn get_withdrawal_request_id(id: &str) -> String {
    format!("withdrawal/{}", id)
}
