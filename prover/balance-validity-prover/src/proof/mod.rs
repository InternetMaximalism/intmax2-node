use anyhow::Context as _;
use base64::prelude::*;
use intmax2_zkp::{
    circuits::{
        balance::balance_processor::BalanceProcessor, validity::validity_circuit::ValidityCircuit,
    },
    common::witness::{
        receive_deposit_witness::ReceiveDepositWitness, receive_transfer_witness::ReceiveTransferWitness, send_witness::SendWitness, update_witness::UpdateWitness
    },
    ethereum_types::u256::U256,
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

pub async fn generate_receive_deposit_proof_job(
    request_id: String,
    public_key: U256,
    prev_balance_proof: Option<ProofWithPublicInputs<F, C, D>>,
    receive_deposit_witness: &ReceiveDepositWitness,
    balance_processor: &BalanceProcessor<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let balance_circuit_data = balance_processor.balance_circuit.data.verifier_data();

    println!("Proving...");
    let balance_proof = balance_processor.prove_receive_deposit(
        public_key,
        &receive_deposit_witness,
        &prev_balance_proof,
    );

    let compressed_balance_proof = balance_proof
        .clone()
        .compress(
            &balance_circuit_data.verifier_only.circuit_digest,
            &balance_circuit_data.common,
        )
        .with_context(|| "Failed to compress proof")?;

    let encoded_compressed_balance_proof =
        BASE64_STANDARD.encode(&compressed_balance_proof.to_bytes());

    let opts = SetOptions::default()
        .conditional_set(ExistenceCheck::NX)
        .get(true)
        .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

    let _ = redis::Cmd::set_options(&request_id, encoded_compressed_balance_proof.clone(), opts)
        .query_async::<_, Option<String>>(conn)
        .await
        .with_context(|| "Failed to set proof")?;

    Ok(())
}

pub async fn generate_balance_update_proof_job(
    request_id: String,
    public_key: U256,
    prev_balance_proof: Option<ProofWithPublicInputs<F, C, D>>,
    balance_update_witness: &UpdateWitness<F, C, D>,
    balance_processor: &BalanceProcessor<F, C, D>,
    validity_circuit: &ValidityCircuit<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let balance_circuit_data = balance_processor.balance_circuit.data.verifier_data();
    // let validity_circuit_data = validity_circuit_data.verifier_data();

    println!("Proving...");
    let balance_proof = balance_processor.prove_update(
        &validity_circuit,
        public_key,
        &balance_update_witness,
        &prev_balance_proof,
    );

    let compressed_balance_proof = balance_proof
        .clone()
        .compress(
            &balance_circuit_data.verifier_only.circuit_digest,
            &balance_circuit_data.common,
        )
        .with_context(|| "Failed to compress proof")?;

    let encoded_compressed_balance_proof =
        BASE64_STANDARD.encode(&compressed_balance_proof.to_bytes());

    let opts = SetOptions::default()
        .conditional_set(ExistenceCheck::NX)
        .get(true)
        .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

    let _ = redis::Cmd::set_options(&request_id, encoded_compressed_balance_proof.clone(), opts)
        .query_async::<_, Option<String>>(conn)
        .await
        .with_context(|| "Failed to set proof")?;

    Ok(())
}

pub async fn generate_balance_transfer_proof_job(
    request_id: String,
    public_key: U256,
    prev_balance_proof: Option<ProofWithPublicInputs<F, C, D>>,
    receive_transfer_witness: &ReceiveTransferWitness<F, C, D>,
    balance_processor: &BalanceProcessor<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let balance_circuit_data = balance_processor.balance_circuit.data.verifier_data();
    // let validity_circuit_data = validity_circuit_data.verifier_data();

    println!("Proving...");
    let balance_proof = balance_processor.prove_receive_transfer(
        public_key,
        receive_transfer_witness,
        &prev_balance_proof,
    );

    let compressed_balance_proof = balance_proof
        .clone()
        .compress(
            &balance_circuit_data.verifier_only.circuit_digest,
            &balance_circuit_data.common,
        )
        .with_context(|| "Failed to compress proof")?;

    let encoded_compressed_balance_proof =
        BASE64_STANDARD.encode(&compressed_balance_proof.to_bytes());

    let opts = SetOptions::default()
        .conditional_set(ExistenceCheck::NX)
        .get(true)
        .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

    let _ = redis::Cmd::set_options(&request_id, encoded_compressed_balance_proof.clone(), opts)
        .query_async::<_, Option<String>>(conn)
        .await
        .with_context(|| "Failed to set proof")?;

    Ok(())
}

pub async fn generate_balance_send_proof_job(
    request_id: String,
    public_key: U256,
    prev_balance_proof: Option<ProofWithPublicInputs<F, C, D>>,
    send_witness: &SendWitness,
    balance_update_witness: &UpdateWitness<F, C, D>,
    balance_processor: &BalanceProcessor<F, C, D>,
    validity_circuit: &ValidityCircuit<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let balance_circuit_data = balance_processor.balance_circuit.data.verifier_data();
    // let validity_circuit_data = validity_circuit_data.verifier_data();

    println!("Proving...");
    let balance_proof = balance_processor.prove_send(
        &validity_circuit,
        public_key,
        &send_witness,
        &balance_update_witness,
        &prev_balance_proof,
    );

    let encoded_compressed_balance_proof = encode_plonky2_proof(balance_proof, &balance_circuit_data);

    let opts = SetOptions::default()
        .conditional_set(ExistenceCheck::NX)
        .get(true)
        .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

    let _ = redis::Cmd::set_options(&request_id, encoded_compressed_balance_proof.clone(), opts)
        .query_async::<_, Option<String>>(conn)
        .await
        .with_context(|| "Failed to set proof")?;

    Ok(())
}
