use anyhow::Context as _;
use intmax2_zkp::{
    circuits::balance::{
        balance_pis::BalancePublicInputs,
        send::spent_circuit::{SpentCircuit, SpentValue},
        transition::transition_processor::BalanceTransitionProcessor,
    },
    common::witness::{
        receive_deposit_witness::ReceiveDepositWitness,
        receive_transfer_witness::ReceiveTransferWitness, send_witness::SendWitness,
        update_witness::UpdateWitness,
    },
};
use plonky2::plonk::{
    circuit_data::{VerifierCircuitData, VerifierOnlyCircuitData},
    config::{GenericConfig, PoseidonGoldilocksConfig},
};
use redis::{ExistenceCheck, SetExpiry, SetOptions};

use crate::app::{config, encode::encode_plonky2_proof};

pub mod serializer;

const D: usize = 2;
type C = PoseidonGoldilocksConfig;
type F = <C as GenericConfig<D>>::F;

pub fn generate_deposit_transition_proof_job(
    prev_balance_pis: &BalancePublicInputs,
    receive_deposit_witness: &ReceiveDepositWitness,
    balance_transition_processor: &BalanceTransitionProcessor<F, C, D>,
    balance_verifier_data: &VerifierCircuitData<F, C, D>,
) -> anyhow::Result<String> {
    log::debug!("Proving...");
    let balance_transition_proof = balance_transition_processor.prove_receive_deposit(
        balance_verifier_data,
        prev_balance_pis,
        receive_deposit_witness,
    );

    let encoded_compressed_balance_transition_proof = encode_plonky2_proof(
        balance_transition_proof,
        &balance_transition_processor
            .balance_transition_circuit
            .data
            .verifier_data(),
    );

    Ok(encoded_compressed_balance_transition_proof)
}

pub fn generate_transfer_transition_proof_job(
    prev_balance_pis: &BalancePublicInputs,
    receive_transfer_witness: &ReceiveTransferWitness<F, C, D>,
    balance_transition_processor: &BalanceTransitionProcessor<F, C, D>,
    balance_verifier_data: &VerifierCircuitData<F, C, D>,
) -> anyhow::Result<String> {
    log::debug!("Proving...");
    let balance_transition_proof = balance_transition_processor.prove_receive_transfer(
        balance_verifier_data,
        &prev_balance_pis,
        receive_transfer_witness,
    );

    let encoded_compressed_balance_transition_proof = encode_plonky2_proof(
        balance_transition_proof,
        &balance_transition_processor
            .balance_transition_circuit
            .data
            .verifier_data(),
    );

    Ok(encoded_compressed_balance_transition_proof)
}

pub fn generate_send_transition_proof_job(
    send_witness: &SendWitness,
    update_witness: &UpdateWitness<F, C, D>,
    balance_transition_processor: &BalanceTransitionProcessor<F, C, D>,
    balance_circuit_vd: &VerifierOnlyCircuitData<C, D>,
    validity_verifier_data: &VerifierCircuitData<F, C, D>,
) -> anyhow::Result<String> {
    log::debug!("Proving...");
    let balance_transition_proof = balance_transition_processor.prove_send(
        &validity_verifier_data,
        balance_circuit_vd,
        send_witness,
        update_witness,
    );

    let encoded_compressed_balance_transition_proof = encode_plonky2_proof(
        balance_transition_proof,
        &balance_transition_processor
            .balance_transition_circuit
            .data
            .verifier_data(),
    );

    Ok(encoded_compressed_balance_transition_proof)
}

pub async fn generate_spent_transition_proof_job(
    request_id: String,
    send_witness: &SendWitness,
    spent_circuit: &SpentCircuit<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let spent_value = SpentValue::new(
        &send_witness.prev_private_state,
        &send_witness.prev_balances,
        send_witness.new_private_state_salt,
        &send_witness.transfers,
        &send_witness.asset_merkle_proofs,
        send_witness.tx_witness.tx.nonce,
    );
    log::debug!("Proving...");
    let spent_proof = spent_circuit.prove(&spent_value).unwrap();

    let encoded_compressed_spent_proof =
        encode_plonky2_proof(spent_proof, &spent_circuit.data.verifier_data());

    let opts = SetOptions::default()
        .conditional_set(ExistenceCheck::NX)
        .get(true)
        .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

    let _ = redis::Cmd::set_options(&request_id, encoded_compressed_spent_proof.clone(), opts)
        .query_async::<_, Option<String>>(conn)
        .await
        .with_context(|| "Failed to set proof")?;

    Ok(())
}
