use intmax2_zkp::common::witness::receive_deposit_witness::ReceiveDepositWitness;
use intmax2_zkp::common::witness::send_witness::SendWitness;
use serde::Deserialize;
use serde::Serialize;

use crate::proof::SerializableReceiveTransferWitness;
use crate::proof::SerializableUpdateWitness;

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

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofResponse {
    pub success: bool,
    pub request_id: String,
    pub proof: Option<String>,
    pub error_message: Option<String>,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofDepositRequest {
    pub prev_balance_proof: Option<String>,
    pub receive_deposit_witness: ReceiveDepositWitness,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct DepositHashQuery {
    pub deposit_hashes: Vec<String>,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofDepositValue {
    pub deposit_hash: String,
    pub proof: String,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofsDepositResponse {
    pub success: bool,
    pub proofs: Vec<ProofDepositValue>,
    pub error_message: Option<String>,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofUpdateRequest {
    pub prev_balance_proof: Option<String>,
    pub balance_update_witness: SerializableUpdateWitness,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct UpdateIdQuery {
    pub block_hashes: Vec<String>,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofUpdateValue {
    pub block_hash: String,
    pub proof: String,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofsUpdateResponse {
    pub success: bool,
    pub proofs: Vec<ProofUpdateValue>,
    pub error_message: Option<String>,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofTransferRequest {
    pub prev_balance_proof: Option<String>,
    pub receive_transfer_witness: SerializableReceiveTransferWitness,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct TransferIdQuery {
    pub private_commitments: Vec<String>,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofTransferValue {
    pub private_commitment: String,
    pub proof: String,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofsTransferResponse {
    pub success: bool,
    pub proofs: Vec<ProofTransferValue>,
    pub error_message: Option<String>,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofSendRequest {
    pub prev_balance_proof: Option<String>,
    pub send_witness: SendWitness,
    pub balance_update_witness: SerializableUpdateWitness,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SendIdQuery {
    pub block_hashes: Vec<String>,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofSendValue {
    pub block_hash: String,
    pub proof: String,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofsSendResponse {
    pub success: bool,
    pub proofs: Vec<ProofSendValue>,
    pub error_message: Option<String>,
}
