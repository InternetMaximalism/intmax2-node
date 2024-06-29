use crate::app::interface::{ErrorResponse, HealthCheckResponse};
use actix_web::{get, web, HttpResponse, Responder, Result};
use std::time::{Duration, SystemTime, UNIX_EPOCH};

#[get("/health")]
async fn health_check(redis: web::Data<redis::Client>) -> Result<impl Responder> {
    let start_time = SystemTime::now();

    let mut conn = redis
        .get_async_connection()
        .await
        .map_err(actix_web::error::ErrorInternalServerError)?;

    let pong: redis::RedisResult<String> = redis::cmd("PING").query_async(&mut conn).await;

    match pong {
        Ok(response) if response == "PONG" => {
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
