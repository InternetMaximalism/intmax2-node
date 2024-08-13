use intmax2_zkp::common::witness::receive_deposit_witness::ReceiveDepositWitness;
use serde::Deserialize;
use serde::Serialize;

#[derive(Debug, Serialize)]
pub struct HealthCheckResponse {
    pub message: String,
    pub timestamp: u128,
    pub uptime: f64,
}

#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub success: bool,
    pub code: u16,
    pub message: String,
}

#[derive(Debug, Deserialize)]
pub struct ProofRequest {
    #[serde(rename = "prevBalanceProof")]
    pub prev_balance_proof: Option<String>,
    #[serde(rename = "receiveDepositWitness")]
    pub receive_deposit_witness: ReceiveDepositWitness,
}

#[derive(Debug, Deserialize)]
pub struct DepositIndexQuery {
    #[serde(rename = "depositIndices")]
    pub deposit_indices: Vec<String>,
}

#[derive(Debug, Serialize)]
pub struct ProofResponse {
    pub success: bool,
    pub proof: Option<String>,
    #[serde(rename = "errorMessage")]
    pub error_message: Option<String>,
}

#[derive(Debug, Serialize)]
pub struct ProofValue {
    #[serde(rename = "depositIndex")]
    pub deposit_index: String,
    pub proof: String,
}

#[derive(Debug, Serialize)]
pub struct ProofsResponse {
    pub success: bool,
    pub proofs: Vec<ProofValue>,
    #[serde(rename = "errorMessage")]
    pub error_message: Option<String>,
}
