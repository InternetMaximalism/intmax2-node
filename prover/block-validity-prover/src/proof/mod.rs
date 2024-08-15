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

use crate::app::{config, encode::encode_plonky2_proof};

const D: usize = 2;
type C = PoseidonGoldilocksConfig;
type F = <C as GenericConfig<D>>::F;

pub async fn generate_block_validity_proof_job(
    request_id: String,
    prev_validity_proof: Option<ProofWithPublicInputs<F, C, D>>,
    validity_witness: ValidityWitness,
    validity_processor: &ValidityProcessor<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let validity_circuit_data = validity_processor
        .validity_circuit
        .data
        .verifier_data();

    log::info!("Proving...");
    let validity_proof = validity_processor
        .prove(&prev_validity_proof, &validity_witness)
        .with_context(|| "Failed to prove block validity")?;

    let encoded_compressed_validity_proof = encode_plonky2_proof(validity_proof, &validity_circuit_data);

    let opts = SetOptions::default()
        .conditional_set(ExistenceCheck::NX)
        .get(true)
        .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

    let _ = redis::Cmd::set_options(&request_id, encoded_compressed_validity_proof.clone(), opts)
        .query_async::<_, Option<String>>(conn)
        .await
        .with_context(|| "Failed to set proof")?;

    Ok(())
}
