use intmax2_zkp::common::witness::compressed::compressed_validity_witness::CompressedValidityWitness;
use intmax2_zkp::common::witness::validity_witness::ValidityWitness;
use serde::Deserialize;
use serde::Serialize;

#[derive(Debug, Serialize)]
pub struct HealthCheckResponse {
    pub message: String,
    pub timestamp: u128,
    pub uptime: f64,
}

#[derive(Debug, Deserialize)]
pub struct ProofRequest {
    #[serde(rename = "blockHash")]
    pub block_hash: String,
    #[serde(rename = "prevValidityProof")]
    pub prev_validity_proof: Option<String>,
    #[serde(rename = "validityWitness")]
    pub validity_witness: Option<CompressedValidityWitness>,
    #[serde(rename = "plainValidityWitness")]
    pub plain_validity_witness: Option<ValidityWitness>,
}

#[derive(Debug, Deserialize)]
pub struct BlockHashQuery {
    #[serde(rename = "blockHashes")]
    pub block_hashes: Vec<String>,
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
    #[serde(rename = "blockHash")]
    pub block_hash: String,
    pub proof: String,
}

#[derive(Debug, Serialize)]
pub struct ProofsResponse {
    pub success: bool,
    pub proofs: Vec<ProofValue>,
    #[serde(rename = "errorMessage")]
    pub error_message: Option<String>,
}
