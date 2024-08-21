use crate::app::{
    interface::{ErrorResponse, HealthCheckResponse},
    state::AppState,
};
use actix_web::{get, web, HttpResponse, Responder, Result};
use std::time::{Duration, SystemTime, UNIX_EPOCH};

#[get("/health")]
async fn health_check(
    redis: web::Data<redis::Client>,
    state: web::Data<AppState>,
) -> Result<impl Responder> {
    let start_time = SystemTime::now();

    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let pong: redis::RedisResult<String> = redis::cmd("PING").query_async(&mut conn).await;
    let is_balance_transition_circuit_built = state.balance_transition_processor.get().is_some();
    let is_balance_circuit_built = state.balance_verifier_data.get().is_some();
    let is_validity_circuit_built = state.validity_verifier_data.get().is_some();

    match pong {
        Ok(response) if response == "PONG" => {
            if !is_balance_transition_circuit_built
                || !is_balance_circuit_built
                || !is_validity_circuit_built
            {
                return Ok(HttpResponse::InternalServerError().json(ErrorResponse {
                    success: false,
                    code: 500,
                    message: "Circuits are not built".to_string(),
                }));
            }

            let message = "OK";
            let timestamp = SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .expect("Time went backwards")
                .as_millis();
            let end_time = SystemTime::now();
            let uptime = end_time
                .duration_since(start_time)
                .unwrap_or(Duration::from_secs(0))
                .as_secs_f64();

            let response = HealthCheckResponse {
                message: message.to_string(),
                timestamp,
                uptime,
            };

            Ok(HttpResponse::Ok().json(response))
        }
        _ => Ok(HttpResponse::InternalServerError().json(ErrorResponse {
            success: false,
            code: 500,
            message: "Failed to receive PONG response".to_string(),
        })),
    }
}
