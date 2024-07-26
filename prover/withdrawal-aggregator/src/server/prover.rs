use crate::app::{
    self,
    interface::{ProofRequest, IdQuery, ProofResponse,ProofValue, ProofsResponse},
};
use actix_web::{error, get, post, web, HttpResponse, Responder, Result,HttpRequest};
use rand::Rng;
use redis::{ExistenceCheck, SetExpiry, SetOptions};

#[get("/proof/{id}")]
async fn get_proof(
    id: web::Path<String>,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let value: String = redis::Cmd::get(&*id)
        .query_async(&mut conn)
        .await
        .map_err(error::ErrorInternalServerError)?;

    let response = ProofResponse {
        success: true,
        value: value.parse::<isize>().unwrap(),
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
    let ids_query: Result<IdQuery, _> = serde_qs::from_str(query_string);
    let ids: Vec<i32>;

    match ids_query {
        Ok(query) => {
            ids = query.ids;
        }
        Err(e) => {
            eprintln!("Failed to deserialize query: {:?}", e);
            return Ok(HttpResponse::BadRequest().body("Invalid query parameters"));
        }
    }

    let mut values: Vec<ProofValue> = Vec::new();
    for id in &ids {
        let value: String = redis::Cmd::get(id.to_string())
            .query_async(&mut conn)
            .await
            .map_err(actix_web::error::ErrorInternalServerError)?;

        match value.parse::<isize>() {
            Ok(parsed_value) => values.push(ProofValue {
                id: (*id).to_string(),
                value: parsed_value,
            }),
            Err(_) => {
                return Ok(HttpResponse::BadRequest().json(ProofsResponse {
                    success: false,
                    values: Vec::new(),
                    error_message: Some(format!("Failed to parse value for ID: {}", id)),
                }));
            }
        }
    }

    let response = ProofsResponse {
        success: true,
        values,
        error_message: None,
    };

    Ok(HttpResponse::Ok().json(response))
}

#[post("/proof")]
async fn generate_proof(
    req: web::Json<ProofRequest>,
    redis: web::Data<redis::Client>,
) -> Result<impl Responder> {
    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(error::ErrorInternalServerError)?;

    let random_value: isize = rand::thread_rng().gen_range(1..100);

    let opts = SetOptions::default()
        .conditional_set(ExistenceCheck::NX)
        .get(true)
        .with_expiration(SetExpiry::EX(app::config::get("proof_expiration")));

    redis::Cmd::set_options(&req.id, random_value, opts)
        .query_async(&mut conn)
        .await
        .map_err(error::ErrorInternalServerError)?;

    let response = ProofResponse {
        success: true,
        value: random_value,
        error_message: None
    };

    Ok(HttpResponse::Ok().json(response))
}
