use crate::app::encode::encode_plonky2_proof;
use intmax2_zkp::{
    circuits::balance::{
        balance_pis::BalancePublicInputs,
        receive::{
            receive_deposit_circuit::{ReceiveDepositCircuit, ReceiveDepositValue},
            receive_targets::{
                private_state_transition::PrivateStateTransitionValue,
                transfer_inclusion::TransferInclusionValue,
            },
            receive_transfer_circuit::{ReceiveTransferCircuit, ReceiveTransferValue},
        },
        send::spent_circuit::{SpentCircuit, SpentValue},
    },
    common::witness::{
        receive_deposit_witness::ReceiveDepositWitness,
        receive_transfer_witness::ReceiveTransferWitness, send_witness::SendWitness,
    },
    ethereum_types::bytes32::Bytes32,
};
use plonky2::plonk::{
    circuit_data::VerifierCircuitData,
    config::{GenericConfig, PoseidonGoldilocksConfig},
};

pub mod serializer;

const D: usize = 2;
type C = PoseidonGoldilocksConfig;
type F = <C as GenericConfig<D>>::F;

pub fn generate_deposit_transition_proof_job(
    prev_balance_pis: &BalancePublicInputs,
    receive_deposit_witness: &ReceiveDepositWitness,
    receive_deposit_circuit: &ReceiveDepositCircuit<F, C, D>,
) -> anyhow::Result<String> {
    let deposit_witness = receive_deposit_witness.deposit_witness.clone();
    let private_transition_witness = receive_deposit_witness.private_witness.clone();

    // assertion
    let deposit = deposit_witness.deposit.clone();
    let nullifier: Bytes32 = deposit.poseidon_hash().into();
    assert_eq!(nullifier, private_transition_witness.nullifier);
    assert_eq!(deposit.token_index, private_transition_witness.token_index);
    assert_eq!(deposit.amount, private_transition_witness.amount);

    let private_state_transition = PrivateStateTransitionValue::new(
        private_transition_witness.token_index,
        private_transition_witness.amount,
        private_transition_witness.nullifier,
        private_transition_witness.new_salt,
        &private_transition_witness.prev_private_state,
        &private_transition_witness.nullifier_proof,
        &private_transition_witness.prev_asset_leaf,
        &private_transition_witness.asset_merkle_proof,
    );

    let receive_deposit_value = ReceiveDepositValue::new(
        prev_balance_pis.pubkey,
        deposit_witness.deposit_salt,
        deposit_witness.deposit_index,
        &deposit_witness.deposit,
        &deposit_witness.deposit_merkle_proof,
        &prev_balance_pis.public_state,
        &private_state_transition,
    );

    println!("Proving...");
    let receive_deposit_proof = receive_deposit_circuit
        .prove(&receive_deposit_value)
        .unwrap();

    // let balance_transition_proof = balance_transition_processor.prove_receive_deposit(
    //     balance_verifier_data,
    //     prev_balance_pis,
    //     receive_deposit_witness,
    // );

    let res = encode_plonky2_proof(
        receive_deposit_proof,
        &receive_deposit_circuit.data.verifier_data(),
    );

    Ok(res)
}

pub fn generate_transfer_transition_proof_job(
    prev_balance_pis: &BalancePublicInputs,
    receive_transfer_witness: &ReceiveTransferWitness<F, C, D>,
    receive_transfer_circuit: &ReceiveTransferCircuit<F, C, D>,
    balance_verifier_data: &VerifierCircuitData<F, C, D>,
) -> anyhow::Result<String> {
    // assertion
    let transfer = receive_transfer_witness.transfer_witness.transfer;
    let nullifier: Bytes32 = transfer.commitment().into();
    let private_witness = receive_transfer_witness.private_witness.clone();
    assert_eq!(nullifier, private_witness.nullifier);
    assert_eq!(transfer.token_index, private_witness.token_index);
    assert_eq!(transfer.amount, private_witness.amount);

    let private_state_transition = PrivateStateTransitionValue::new(
        private_witness.token_index,
        private_witness.amount,
        private_witness.nullifier,
        private_witness.new_salt,
        &private_witness.prev_private_state,
        &private_witness.nullifier_proof,
        &private_witness.prev_asset_leaf,
        &private_witness.asset_merkle_proof,
    );
    let transfer_witness = receive_transfer_witness.transfer_witness.clone();
    let transfer_inclusion = TransferInclusionValue::new(
        balance_verifier_data,
        &transfer,
        transfer_witness.transfer_index,
        &transfer_witness.transfer_merkle_proof,
        &transfer_witness.tx,
        &receive_transfer_witness.balance_proof,
    );
    let receive_transfer_value = ReceiveTransferValue::new(
        &prev_balance_pis.public_state,
        &receive_transfer_witness.block_merkle_proof,
        &transfer_inclusion,
        &private_state_transition,
    );

    println!("Proving...");
    let receive_transfer_proof = receive_transfer_circuit
        .prove(&receive_transfer_value)
        .unwrap();

    // let balance_transition_proof = balance_transition_processor.prove_receive_transfer(
    //     balance_verifier_data,
    //     prev_balance_pis,
    //     receive_transfer_witness,
    // );

    let res = encode_plonky2_proof(
        receive_transfer_proof,
        &receive_transfer_circuit.data.verifier_data(),
    );

    Ok(res)
}

pub fn generate_spent_proof_job(
    send_witness: &SendWitness,
    spent_circuit: &SpentCircuit<F, C, D>,
) -> anyhow::Result<String> {
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

    let res = encode_plonky2_proof(spent_proof, &spent_circuit.data.verifier_data());

    Ok(res)
}
