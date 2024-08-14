use serde::{Deserialize, Serialize};

use super::encode::SerializableWithdrawalWitness;

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
pub struct ProofRequest {
    pub id: String,
    pub prev_withdrawal_proof: Option<String>,
    pub withdrawal_witness: SerializableWithdrawalWitness,
}

#[derive(Deserialize)]
pub struct IdQuery {
    pub ids: Vec<i32>,
}

#[derive(Serialize)]
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
