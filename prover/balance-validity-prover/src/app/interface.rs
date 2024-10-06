use intmax2_zkp::common::insufficient_flags::InsufficientFlags;
use intmax2_zkp::common::private_state::PrivateState;
use intmax2_zkp::common::salt::Salt;
use intmax2_zkp::common::transfer::Transfer;
use intmax2_zkp::common::trees::asset_tree::AssetLeaf;
use intmax2_zkp::common::trees::asset_tree::AssetMerkleProof;
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
    pub request_id: String,
    pub prev_balance_proof: Option<String>,
    pub receive_deposit_witness: ReceiveDepositWitness,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct DepositHashQuery {
    pub request_ids: Vec<String>,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofDepositValue {
    pub request_id: String,
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
    pub request_id: String,
    pub prev_balance_proof: Option<String>,
    pub balance_update_witness: SerializableUpdateWitness,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct UpdateIdQuery {
    pub request_ids: Vec<String>,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofUpdateValue {
    pub request_id: String,
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
    pub request_id: String,
    pub prev_balance_proof: Option<String>,
    pub receive_transfer_witness: SerializableReceiveTransferWitness,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct TransferIdQuery {
    pub request_ids: Vec<String>,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofTransferValue {
    pub request_id: String,
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
    pub request_id: String,
    pub prev_balance_proof: Option<String>,
    pub send_witness: SendWitness,
    pub balance_update_witness: SerializableUpdateWitness,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SendIdQuery {
    pub request_ids: Vec<String>,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofSendValue {
    pub request_id: String,
    pub proof: String,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofsSendResponse {
    pub success: bool,
    pub proofs: Vec<ProofSendValue>,
    pub error_message: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SpentTokenWitness {
    pub prev_private_state: PrivateState,
    pub prev_balances: Vec<AssetLeaf>,
    pub asset_merkle_proofs: Vec<AssetMerkleProof>,
    pub insufficient_flags: InsufficientFlags,
    pub transfers: Vec<Transfer>,
    pub new_private_state_salt: Salt,
    pub tx_nonce: u32,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofSpendRequest {
    pub request_id: String,
    pub send_witness: SpentTokenWitness, // TODO: rename to spent_token_witness
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SpentIdQuery {
    pub request_ids: Vec<String>,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofSpendValue {
    pub request_id: String,
    pub proof: String,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProofsSpentResponse {
    pub success: bool,
    pub proofs: Vec<ProofSpendValue>,
    pub error_message: Option<String>,
}
