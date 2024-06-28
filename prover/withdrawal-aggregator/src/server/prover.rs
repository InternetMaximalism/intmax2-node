use crate::app::{
    self,
    interface::{ProofRequest, ProofResponse},
};
use actix_web::{error, get, post, web, HttpResponse, Responder, Result};
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
    };

    Ok(HttpResponse::Ok().json(response))
}
