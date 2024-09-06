use crate::app::config;
use anyhow::Context;
use intmax2_zkp::{
    circuits::fraud::fraud_processor::FraudProcessor, ethereum_types::address::Address,
};
use plonky2::{
    field::goldilocks_field::GoldilocksField,
    plonk::{config::PoseidonGoldilocksConfig, proof::ProofWithPublicInputs},
};
use redis::{ExistenceCheck, SetExpiry, SetOptions};

const D: usize = 2;
type C = PoseidonGoldilocksConfig;
type F = GoldilocksField;

pub async fn generate_fraud_proof_job(
    request_id: String,
    challenger: Address,
    validity_proof: &ProofWithPublicInputs<F, C, D>,
    fraud_processor: &FraudProcessor,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    log::debug!("Proving...");
    let fraud_proof = fraud_processor
        .prove(challenger, validity_proof)
        .with_context(|| "Failed to prove fraud")?;

    let encoded_compressed_fraud_proof =
        serde_json::to_string(&fraud_proof).with_context(|| "Failed to encode fraud proof")?;

    let opts = SetOptions::default()
        .conditional_set(ExistenceCheck::NX)
        .get(true)
        .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

    let _ = redis::Cmd::set_options(&request_id, encoded_compressed_fraud_proof.clone(), opts)
        .query_async::<_, Option<String>>(conn)
        .await
        .with_context(|| "Failed to set proof")?;

    Ok(())
}
