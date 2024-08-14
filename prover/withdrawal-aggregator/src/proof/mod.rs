use crate::app::{config, encode::encode_plonky2_proof};
use anyhow::Context;
use intmax2_zkp::{
    circuits::withdrawal::withdrawal_processor::WithdrawalProcessor,
    common::witness::withdrawal_witness::WithdrawalWitness,
};
use plonky2::{
    field::goldilocks_field::GoldilocksField,
    plonk::{config::PoseidonGoldilocksConfig, proof::ProofWithPublicInputs},
};
use redis::{ExistenceCheck, SetExpiry, SetOptions};

const D: usize = 2;
type C = PoseidonGoldilocksConfig;
type F = GoldilocksField;

pub async fn generate_withdrawal_proof_job(
    request_id: String,
    prev_withdrawal_proof: Option<ProofWithPublicInputs<F, C, D>>,
    withdrawal_witness: &WithdrawalWitness<F, C, D>,
    withdrawal_processor: &WithdrawalProcessor<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let balance_circuit_data = withdrawal_processor.withdrawal_circuit.data.verifier_data();

    log::info!("Proving...");
    let withdrawal_proof = withdrawal_processor
        .prove(withdrawal_witness, &prev_withdrawal_proof)
        .with_context(|| "Failed to prove withdrawal")?;

    let encoded_compressed_withdrawal_proof =
        encode_plonky2_proof(withdrawal_proof, &balance_circuit_data);

    let opts = SetOptions::default()
        .conditional_set(ExistenceCheck::NX)
        .get(true)
        .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

    let _ = redis::Cmd::set_options(
        &request_id,
        encoded_compressed_withdrawal_proof.clone(),
        opts,
    )
    .query_async::<_, Option<String>>(conn)
    .await
    .with_context(|| "Failed to set proof")?;

    Ok(())
}
