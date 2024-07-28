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

#[derive(Deserialize)]
pub struct IdQuery {
    pub ids: Vec<i32>,
}

#[derive(Serialize)]
pub struct ProofResponse {
    pub success: bool,
    pub value: isize,
    pub error_message: Option<String>,
}

#[derive(Serialize)]
pub struct ProofValue {
    pub id: String,
    pub value: isize,
}

#[derive(Serialize)]
pub struct ProofsResponse {
    pub success: bool,
    pub values: Vec<ProofValue>,
    pub error_message: Option<String>,
}
