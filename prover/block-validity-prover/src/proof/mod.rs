use anyhow::Context as _;
use base64::prelude::*;
use intmax2_zkp::{
    circuits::validity::validity_processor::ValidityProcessor,
    common::witness::validity_witness::ValidityWitness,
};
use plonky2::plonk::{
    config::{GenericConfig, PoseidonGoldilocksConfig},
    proof::ProofWithPublicInputs,
};
use redis::{ExistenceCheck, SetExpiry, SetOptions};

use crate::app::config;

const D: usize = 2;
type C = PoseidonGoldilocksConfig;
type F = <C as GenericConfig<D>>::F;

pub async fn generate_proof_job(
    request_id: String,
    prev_validity_proof: Option<ProofWithPublicInputs<F, C, D>>,
    validity_witness: ValidityWitness,
    validity_processor: &ValidityProcessor<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let validity_circuit = validity_processor
        .validity_circuit
        .data
        .verifier_data();

    println!("Proving...");
    let validity_proof = validity_processor
        .prove(&prev_validity_proof, &validity_witness)
        .with_context(|| "Failed to prove block validity")?;

    let compressed_validity_proof = validity_proof
        .clone()
        .compress(
            &validity_circuit.verifier_only.circuit_digest,
            &validity_circuit.common,
        )
        .with_context(|| "Failed to compress proof")?;

    let encoded_compressed_validity_proof =
        BASE64_STANDARD.encode(&compressed_validity_proof.to_bytes());

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
