use serde::{Deserialize, Serialize};

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

#[derive(Clone, Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct FraudProofRequest {
    pub validity_proof: String,
    pub challenger: String,
}

#[derive(Deserialize)]
pub struct FraudIdQuery {
    pub ids: Vec<String>,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofResponse {
    pub success: bool,
    pub proof: Option<String>,
    pub error_message: Option<String>,
}

#[derive(Serialize)]
pub struct ProofValue {
    pub id: String,
    pub proof: String,
}

#[derive(Serialize)]
pub struct ProofsResponse {
    pub success: bool,
    pub proofs: Vec<ProofValue>,
    pub error_message: Option<String>,
}

#[derive(Serialize)]
pub struct GenerateProofResponse {
    pub success: bool,
    pub message: String,
}
