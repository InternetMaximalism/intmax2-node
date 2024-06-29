use serde::Deserialize;
use serde::Serialize;

#[derive(Serialize)]
pub struct HealthCheckResponse {
    pub message: String,
    pub timestamp: u128,
    pub uptime: f64,
}

#[derive(Serialize)]
pub struct ErrorResponse {
    pub success: bool,
    pub code: u16,
    pub message: String,
}

#[derive(Deserialize)]
pub struct ProofRequest {
    pub id: String,
}

#[derive(Serialize)]
pub struct ProofResponse {
    pub success: bool,
    pub value: isize,
}
