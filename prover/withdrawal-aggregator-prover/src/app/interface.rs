use serde::{Deserialize, Serialize};

#[derive(Serialize)]
pub struct HealthCheckResponse {
    pub message: String,
    pub timestamp: u128,
    pub uptime: f64,
}

// #[derive(Clone, Debug, Deserialize)]
// #[serde(rename_all = "camelCase")]
// pub struct WithdrawalProofRequest {
//     pub id: String,
//     pub prev_withdrawal_proof: Option<String>,
//     pub withdrawal_witness: SerializableWithdrawalWitness,
// }

#[derive(Clone, Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct WithdrawalProofRequest {
    pub id: String,
    pub prev_withdrawal_proof: Option<String>,
    pub single_withdrawal_proof: String,
}

#[derive(Deserialize)]
pub struct WithdrawalIdQuery {
    pub ids: Vec<String>,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct WithdrawalWrapperProofRequest {
    pub id: String,
    pub withdrawal_proof: String,
    pub withdrawal_aggregator: String,
}

#[derive(Deserialize)]
pub struct WithdrawalWrapperIdQuery {
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
