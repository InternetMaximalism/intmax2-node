use anyhow::Context as _;
use intmax2_zkp::{
    circuits::validity::validity_processor::ValidityProcessor,
    common::witness::validity_witness::ValidityWitness,
};
use plonky2::plonk::{
    config::{GenericConfig, PoseidonGoldilocksConfig},
    proof::ProofWithPublicInputs,
};
use redis::{ExistenceCheck, SetExpiry, SetOptions};
use serde::{Deserialize, Serialize};

use crate::app::{config, encode::encode_plonky2_proof};

const D: usize = 2;
type C = PoseidonGoldilocksConfig;
type F = <C as GenericConfig<D>>::F;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RedisResponse {
    pub success: bool,
    pub message: String,
}

pub async fn generate_block_validity_proof_job(
    request_id: String,
    prev_validity_proof: Option<ProofWithPublicInputs<F, C, D>>,
    validity_witness: ValidityWitness,
    validity_processor: &ValidityProcessor<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let result =
        generate_block_validity_proof(prev_validity_proof, validity_witness, validity_processor);

    match result {
        Ok(encoded_compressed_validity_proof) => {
            let response = RedisResponse {
                success: true,
                message: encoded_compressed_validity_proof,
            };
            let response_json = serde_json::to_string(&response)
                .expect("Failed to serialize success response to JSON");
            let opts = SetOptions::default()
                .conditional_set(ExistenceCheck::NX)
                .get(true)
                .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

            let _ = redis::Cmd::set_options(&request_id, response_json, opts)
                .query_async::<_, Option<String>>(conn)
                .await
                .with_context(|| "Failed to set proof")?;

            Ok(())
        }
        Err(e) => {
            log::error!("Failed to generate proof: {:?}", e);
            let response = RedisResponse {
                success: false,
                message: format!("Failed to generate proof: {:?}", e),
            };
            let response_json = serde_json::to_string(&response)
                .expect("Failed to serialize failure response to JSON");
            let opts = SetOptions::default()
                .conditional_set(ExistenceCheck::NX)
                .get(true)
                .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

            let _ = redis::Cmd::set_options(&request_id, response_json, opts)
                .query_async::<_, Option<String>>(conn)
                .await
                .with_context(|| "Failed to set proof")?;

            Err(e)
        }
    }
}

pub fn generate_block_validity_proof(
    prev_validity_proof: Option<ProofWithPublicInputs<F, C, D>>,
    validity_witness: ValidityWitness,
    validity_processor: &ValidityProcessor<F, C, D>,
) -> anyhow::Result<String> {
    let validity_circuit_data = validity_processor.validity_circuit.data.verifier_data();

    log::info!("Proving...");
    let validity_proof = validity_processor
        .prove(&prev_validity_proof, &validity_witness)
        .with_context(|| "Failed to prove block validity")?;

    let encoded_compressed_validity_proof =
        encode_plonky2_proof(validity_proof, &validity_circuit_data)
            .map_err(|e| anyhow::anyhow!("Failed to encode validity proof: {:?}", e))?;

    Ok(encoded_compressed_validity_proof)
}
