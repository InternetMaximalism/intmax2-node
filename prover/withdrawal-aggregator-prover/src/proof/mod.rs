use crate::app::{config, encode::encode_plonky2_proof};
use anyhow::Context;
use intmax2_zkp::{
    circuits::withdrawal::{
        withdrawal_processor::WithdrawalProcessor,
        withdrawal_wrapper_processor::WithdrawalWrapperProcessor,
    },
    common::{withdrawal::Withdrawal, witness::withdrawal_witness::WithdrawalWitness},
    ethereum_types::address::Address,
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
    single_withdrawal_proof: &ProofWithPublicInputs<F, C, D>,
    withdrawal_processor: &WithdrawalProcessor<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let withdrawal_circuit_data = withdrawal_processor.withdrawal_circuit.data.verifier_data();

    withdrawal_processor
        .single_withdrawal_circuit
        .verify(single_withdrawal_proof)
        .map_err(|e| anyhow::anyhow!("Invalid single withdrawal proof: {:?}", e))?;

    log::debug!("Proving...");
    let withdrawal_proof = withdrawal_processor
        .prove_chain(single_withdrawal_proof, &prev_withdrawal_proof)
        .map_err(|e| anyhow::anyhow!("Failed to prove withdrawal chain: {}", e))?;
    let withdrawal = Withdrawal::from_u64_slice(&withdrawal_proof.public_inputs.to_u64_vec());

    let encoded_compressed_withdrawal_proof =
        encode_plonky2_proof(withdrawal_proof, &withdrawal_circuit_data)
            .with_context(|| "Failed to encode withdrawal")?;

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

pub async fn generate_withdrawal_wrapper_proof_job(
    request_id: String,
    withdrawal_proof: ProofWithPublicInputs<F, C, D>,
    withdrawal_aggregator: Address,
    withdrawal_wrapper_processor: &WithdrawalWrapperProcessor,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    log::debug!("Proving...");
    let wrapped_withdrawal_proof = withdrawal_wrapper_processor
        .prove(&withdrawal_proof, withdrawal_aggregator)
        .with_context(|| "Failed to prove withdrawal")?;

    // NOTICE: Not compressing the proof here
    let withdrawal_proof_json = serde_json::to_string(&wrapped_withdrawal_proof)
        .with_context(|| "Failed to encode wrapped withdrawal proof")?;

    let opts = SetOptions::default()
        .conditional_set(ExistenceCheck::NX)
        .get(true)
        .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

    let _ = redis::Cmd::set_options(&request_id, withdrawal_proof_json.clone(), opts)
        .query_async::<_, Option<String>>(conn)
        .await
        .with_context(|| "Failed to set proof")?;

    Ok(())
}
