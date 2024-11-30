use anyhow::Context as _;
use intmax2_zkp::{
    circuits::{
        balance::{
            balance_pis::BalancePublicInputs,
            balance_processor::BalanceProcessor,
            receive::{
                receive_deposit_circuit::ReceiveDepositValue,
                receive_targets::private_state_transition::PrivateStateTransitionValue,
            },
            send::spent_circuit::SpentValue,
        },
        validity::validity_circuit::ValidityCircuit,
        withdrawal::single_withdrawal_circuit::SingleWithdrawalCircuit,
    },
    common::{
        deposit::{get_pubkey_salt_hash, Deposit},
        public_state::PublicState,
        salt::Salt,
        trees::{
            account_tree::AccountMembershipProof, block_hash_tree::BlockHashMerkleProof,
            deposit_tree::DepositMerkleProof,
        },
        witness::{
            private_transition_witness::PrivateTransitionWitness,
            receive_deposit_witness::ReceiveDepositWitness,
            receive_transfer_witness::ReceiveTransferWitness, transfer_witness::TransferWitness,
            tx_witness::TxWitness, update_witness::UpdateWitness,
            withdrawal_witness::WithdrawalWitness,
        },
    },
    ethereum_types::{bytes32::Bytes32, u256::U256, u32limb_trait::U32LimbTrait},
    utils::{leafable::Leafable, recursively_verifiable::RecursivelyVerifiable},
};
use plonky2::plonk::{
    circuit_data::VerifierCircuitData,
    config::{GenericConfig, PoseidonGoldilocksConfig},
    proof::ProofWithPublicInputs,
};
use redis::{ExistenceCheck, SetExpiry, SetOptions};
use serde::{Deserialize, Serialize};

use crate::app::{
    config,
    encode::{decode_plonky2_proof, encode_plonky2_proof},
    interface::SpentWitness,
};

const D: usize = 2;
type C = PoseidonGoldilocksConfig;
type F = <C as GenericConfig<D>>::F;

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SerializableUpdateWitness {
    pub is_prev_account_tree: bool,
    pub validity_proof: String,
    pub block_merkle_proof: BlockHashMerkleProof,
    pub account_membership_proof: AccountMembershipProof,
}

impl SerializableUpdateWitness {
    pub fn decode(
        &self,
        balance_circuit_data: &VerifierCircuitData<F, C, D>,
    ) -> anyhow::Result<UpdateWitness<F, C, D>> {
        let validity_proof = decode_plonky2_proof(&self.validity_proof, balance_circuit_data)?;
        Ok(UpdateWitness {
            is_prev_account_tree: self.is_prev_account_tree,
            validity_proof,
            block_merkle_proof: self.block_merkle_proof.clone(),
            account_membership_proof: self.account_membership_proof.clone(),
        })
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SerializableReceiveTransferWitness {
    pub transfer_witness: TransferWitness,
    pub private_transition_witness: PrivateTransitionWitness,
    pub sender_balance_proof: String,
    pub block_merkle_proof: BlockHashMerkleProof,
}

impl SerializableReceiveTransferWitness {
    pub fn decode(
        &self,
        balance_circuit_data: &VerifierCircuitData<F, C, D>,
    ) -> anyhow::Result<ReceiveTransferWitness<F, C, D>> {
        let sender_balance_proof =
            decode_plonky2_proof(&self.sender_balance_proof, balance_circuit_data)
                .map_err(|e| anyhow::anyhow!("Failed to decode balance proof: {:?}", e))?;
        Ok(ReceiveTransferWitness {
            transfer_witness: self.transfer_witness.clone(),
            private_transition_witness: self.private_transition_witness.clone(),
            sender_balance_proof,
            block_merkle_proof: self.block_merkle_proof.clone(),
        })
    }
}

// #[derive(Debug, Clone, Serialize, Deserialize)]
// #[serde(rename_all = "camelCase")]
// pub struct SerializableWithdrawalWitness {
//     pub transfer_witness: TransferWitness,
//     pub balance_proof: String,
// }

// impl SerializableWithdrawalWitness {
//     pub fn decode(
//         &self,
//         balance_circuit_data: &VerifierCircuitData<F, C, D>,
//     ) -> anyhow::Result<WithdrawalWitness<F, C, D>> {
//         let balance_proof = decode_plonky2_proof(&self.balance_proof, balance_circuit_data)
//             .map_err(|e| anyhow::anyhow!("Failed to decode balance proof: {:?}", e))?;
//         Ok(WithdrawalWitness {
//             transfer_witness: self.transfer_witness.clone(),
//             balance_proof,
//         })
//     }
// }

pub async fn generate_receive_deposit_proof_job(
    full_request_id: String,
    public_key: U256,
    prev_balance_proof: Option<ProofWithPublicInputs<F, C, D>>,
    receive_deposit_witness: &ReceiveDepositWitness,
    balance_processor: &BalanceProcessor<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let balance_circuit_data = balance_processor.balance_circuit.data.verifier_data();

    if let Some(prev_balance_proof) = &prev_balance_proof {
        let prev_balance_proof = BalancePublicInputs::from_pis(&prev_balance_proof.public_inputs);
        println!(
            "prev_balance_proof account_tree_root: {}",
            prev_balance_proof
                .public_state
                .account_tree_root
                .to_string()
        );
        println!(
            "prev_balance_proof private_commitment: {}",
            prev_balance_proof.private_commitment.to_string()
        );
    }

    log::debug!("Proving...");
    let balance_proof = balance_processor
        .prove_receive_deposit(public_key, &receive_deposit_witness, &prev_balance_proof)
        .with_context(|| "Failed to prove receive deposit")?;

    let balance_pis = BalancePublicInputs::from_pis(&balance_proof.public_inputs);
    println!("balance_proof: {:?}", balance_pis);
    println!(
        "balance_proof private_commitment: {}",
        balance_pis.private_commitment.to_string()
    );
    println!(
        "balance_proof account_tree_root: {}",
        balance_pis.public_state.account_tree_root.to_string()
    );
    println!(
        "balance_proof prev_account_tree_root: {}",
        balance_pis.public_state.prev_account_tree_root.to_string()
    );

    let encoded_compressed_balance_proof =
        encode_plonky2_proof(balance_proof.clone(), &balance_circuit_data)?;

    let opts = SetOptions::default()
        .conditional_set(ExistenceCheck::NX)
        .get(true)
        .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

    let _ = redis::Cmd::set_options(
        &full_request_id,
        encoded_compressed_balance_proof.clone(),
        opts,
    )
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
) -> anyhow::Result<String> {
    let balance_circuit_data = balance_processor.balance_circuit.data.verifier_data();
    // let validity_circuit_data = validity_circuit_data.verifier_data();

    log::debug!("Proving...");
    let balance_proof = balance_processor
        .prove_update(
            &validity_circuit.data.verifier_data(),
            public_key,
            &balance_update_witness,
            &prev_balance_proof,
        )
        .with_context(|| "Failed to prove update")?;

    let encoded_compressed_balance_proof =
        encode_plonky2_proof(balance_proof.clone(), &balance_circuit_data)?;

    let opts = SetOptions::default()
        .conditional_set(ExistenceCheck::NX)
        .get(true)
        .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

    let _ = redis::Cmd::set_options(&request_id, encoded_compressed_balance_proof.clone(), opts)
        .query_async::<_, Option<String>>(conn)
        .await
        .with_context(|| "Failed to set proof")?;

    Ok(encoded_compressed_balance_proof)
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

    log::debug!("Proving...");
    let balance_proof = balance_processor
        .prove_receive_transfer(public_key, receive_transfer_witness, &prev_balance_proof)
        .with_context(|| "Failed to prove receive transfer")?;

    let encoded_compressed_balance_proof =
        encode_plonky2_proof(balance_proof.clone(), &balance_circuit_data)?;

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
    full_request_id: String,
    public_key: U256,
    prev_balance_proof: Option<ProofWithPublicInputs<F, C, D>>,
    tx_witness: &TxWitness,
    spent_proof: &ProofWithPublicInputs<F, C, D>,
    balance_update_witness: &UpdateWitness<F, C, D>,
    balance_processor: &BalanceProcessor<F, C, D>,
    validity_circuit: &ValidityCircuit<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let balance_circuit_data = balance_processor.balance_circuit.data.verifier_data();
    // let validity_circuit_data = validity_circuit_data.verifier_data();

    if let Some(prev_balance_proof) = &prev_balance_proof {
        let prev_balance_proof = BalancePublicInputs::from_pis(&prev_balance_proof.public_inputs);
        println!(
            "prev_balance_proof: {}",
            prev_balance_proof
                .public_state
                .account_tree_root
                .to_string()
        );
    }

    log::debug!("Proving...");
    let balance_proof = balance_processor
        .prove_send(
            &validity_circuit.data.verifier_data(),
            public_key,
            tx_witness,
            &balance_update_witness,
            spent_proof,
            &prev_balance_proof,
        )
        .with_context(|| "Failed to prove send")?;

    let balance_pis = BalancePublicInputs::from_pis(&balance_proof.public_inputs);
    println!("balance_proof: {:?}", balance_pis);
    println!(
        "balance_proof account_tree_root: {}",
        balance_pis.public_state.account_tree_root.to_string()
    );
    println!(
        "balance_proof prev_account_tree_root: {}",
        balance_pis.public_state.prev_account_tree_root.to_string()
    );

    let encoded_compressed_balance_proof =
        encode_plonky2_proof(balance_proof, &balance_circuit_data)
            .map_err(|e| anyhow::anyhow!("Failed to encode balance proof: {:?}", e))?;

    let opts = SetOptions::default()
        .conditional_set(ExistenceCheck::NX)
        .get(true)
        .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

    let _ = redis::Cmd::set_options(
        &full_request_id,
        encoded_compressed_balance_proof.clone(),
        opts,
    )
    .query_async::<_, Option<String>>(conn)
    .await
    .with_context(|| "Failed to set proof")?;

    Ok(())
}

pub fn generate_balance_spend_proof_job(
    spent_token_witness: &SpentWitness,
    balance_processor: &BalanceProcessor<F, C, D>,
) -> anyhow::Result<String> {
    let spent_circuit = &balance_processor
        .balance_transition_processor
        .sender_processor
        .spent_circuit;

    let spent_value = SpentValue::new(
        &spent_token_witness.prev_private_state,
        &spent_token_witness.prev_balances,
        spent_token_witness.new_private_state_salt,
        &spent_token_witness.transfers,
        &spent_token_witness.asset_merkle_proofs,
        spent_token_witness.tx.nonce,
    )
    .with_context(|| "Failed to create spent value")?;

    log::debug!("Proving...");
    let spent_proof = spent_circuit.prove(&spent_value).unwrap();

    let encoded_compressed_spent_proof =
        encode_plonky2_proof(spent_proof, &spent_circuit.data.verifier_data())
            .map_err(|e| anyhow::anyhow!("Failed to encode balance proof: {:?}", e))?;

    Ok(encoded_compressed_spent_proof)
}

pub fn generate_balance_withdrawal_proof_job(
    withdrawal_witness: &WithdrawalWitness<F, C, D>,
    balance_processor: &BalanceProcessor<F, C, D>,
    single_withdrawal_circuit: &SingleWithdrawalCircuit<F, C, D>,
) -> anyhow::Result<String> {
    let transition_inclusion_value = withdrawal_witness
        .to_transition_inclusion_value(&balance_processor.get_verifier_data())
        .map_err(|e| anyhow::anyhow!("failed to create transition inclusion value: {}", e))?;

    log::debug!("Proving...");
    let single_withdrawal_proof = single_withdrawal_circuit
        .prove(&transition_inclusion_value)
        .map_err(|e| anyhow::anyhow!("failed to prove single withdrawal: {}", e))?;

    let encoded_compressed_single_withdrawal_proof = encode_plonky2_proof(
        single_withdrawal_proof,
        &single_withdrawal_circuit.circuit_data().verifier_data(),
    )
    .map_err(|e| anyhow::anyhow!("Failed to encode balance proof: {:?}", e))?;

    Ok(encoded_compressed_single_withdrawal_proof)
}

// pub fn generate_balance_single_send_proof_job(
//     send_witness: &SendWitness,
//     update_witness: &UpdateWitness<F, C, D>,
//     balance_processor: &BalanceProcessor<F, C, D>,
//     validity_circuit: &ValidityCircuit<F, C, D>,
// ) -> anyhow::Result<String> {
//     let sender_processor = &balance_processor
//         .balance_transition_processor
//         .sender_processor;

//     log::debug!("Proving...");
//     let send_proof = sender_processor.prove(&validity_circuit, &send_witness, &update_witness);

//     let encoded_compressed_send_proof = encode_plonky2_proof(
//         send_proof,
//         &sender_processor.sender_circuit.data.verifier_data(),
//     )
//     .map_err(|e| anyhow::anyhow!("Failed to encode balance proof: {:?}", e))?;

//     Ok(encoded_compressed_send_proof)
// }

pub fn validate_witness(
    _pubkey: U256,
    public_state: &PublicState,
    receive_deposit_witness: &ReceiveDepositWitness,
    prev_balance_proof: &Option<ProofWithPublicInputs<F, C, D>>,
) -> anyhow::Result<()> {
    let deposit_witness = receive_deposit_witness.deposit_witness.clone();
    let private_transition_witness = receive_deposit_witness.private_transition_witness.clone();

    let deposit_index = receive_deposit_witness.deposit_witness.deposit_index;
    let deposit = &receive_deposit_witness.deposit_witness.deposit;
    let deposit_merkle_proof = &receive_deposit_witness.deposit_witness.deposit_merkle_proof;
    println!("siblings: {:?}", deposit_merkle_proof);
    println!("deposit hash: {}", deposit.hash().to_hex());
    println!("deposit index: {}", deposit_index);
    println!(
        "deposit tree root: {}",
        public_state.deposit_tree_root.to_hex()
    );

    // let pubkey_salt_hash = get_pubkey_salt_hash(pubkey, deposit_salt);
    // if pubkey_salt_hash != deposit.pubkey_salt_hash {
    //     anyhow::bail!("pubkey_salt_hash not match");
    // }

    let result =
        deposit_merkle_proof.verify(&deposit, deposit_index, public_state.deposit_tree_root);
    if !result.is_ok() {
        println!("deposit_merkle_proof: {:?}", deposit_merkle_proof);
        anyhow::bail!("Invalid deposit merkle proof");
    }

    let deposit = deposit_witness.deposit.clone();
    let nullifier: Bytes32 = deposit.poseidon_hash().into();
    if nullifier != private_transition_witness.nullifier {
        println!("deposit: {:?}", deposit);
        println!("nullifier: {}", nullifier);
        println!(
            "private_transition_witness.nullifier: {}",
            private_transition_witness.nullifier
        );
        anyhow::bail!("nullifier not match");
    }
    // assert_eq!(deposit.token_index, private_transition_witness.token_index);
    if deposit.token_index != private_transition_witness.token_index {
        println!("token_index: {}", deposit.token_index);
        println!(
            "private_transition_witness.token_index: {}",
            private_transition_witness.token_index
        );
        anyhow::bail!("token_index not match");
    }
    // assert_eq!(deposit.amount, private_transition_witness.amount);
    if deposit.amount != private_transition_witness.amount {
        println!("amount: {}", deposit.amount);
        println!(
            "private_transition_witness.amount: {}",
            private_transition_witness.amount
        );
        anyhow::bail!("amount not match");
    }

    // assertion
    let private_state_transition = PrivateStateTransitionValue::new(
        private_transition_witness.token_index,
        private_transition_witness.amount,
        private_transition_witness.nullifier,
        private_transition_witness.new_salt,
        &private_transition_witness.prev_private_state,
        &private_transition_witness.nullifier_proof,
        &private_transition_witness.prev_asset_leaf,
        &private_transition_witness.asset_merkle_proof,
    )
    .with_context(|| "Failed to create private state transition value")?;

    let prev_balance_pis = if let Some(prev_balance_proof) = prev_balance_proof {
        BalancePublicInputs::from_pis(&prev_balance_proof.public_inputs)
    } else {
        BalancePublicInputs::new(_pubkey)
    };

    let receive_deposit_value = validate_receive_deposit_value(
        prev_balance_pis.pubkey,
        deposit_witness.deposit_salt,
        deposit_witness.deposit_index,
        &deposit_witness.deposit,
        &deposit_witness.deposit_merkle_proof,
        &prev_balance_pis.public_state,
        &private_state_transition,
    )
    .map_err(|e| anyhow::anyhow!("Failed to validate receive deposit value: {:?}", e))?;

    println!(
        "private commitment: {:?}",
        receive_deposit_value.prev_private_commitment
    );
    println!("state: {:?}", prev_balance_pis.private_commitment);

    anyhow::ensure!(
        receive_deposit_value.prev_private_commitment == prev_balance_pis.private_commitment,
        "prev_private_commitment not match"
    );

    Ok(())
}

fn validate_receive_deposit_value(
    pubkey: U256,
    deposit_salt: Salt,
    deposit_index: usize,
    deposit: &Deposit,
    deposit_merkle_proof: &DepositMerkleProof,
    public_state: &PublicState,
    private_state_transition: &PrivateStateTransitionValue,
) -> anyhow::Result<ReceiveDepositValue> {
    // verify deposit inclusion
    let pubkey_salt_hash = get_pubkey_salt_hash(pubkey, deposit_salt);
    anyhow::ensure!(
        pubkey_salt_hash == deposit.pubkey_salt_hash,
        "pubkey_salt_hash not match"
    );
    deposit_merkle_proof
        .verify(&deposit, deposit_index, public_state.deposit_tree_root)
        .map_err(|e| anyhow::anyhow!("Invalid deposit merkle proof: {:?}", e))?;

    let nullifier: Bytes32 = deposit.poseidon_hash().into();
    anyhow::ensure!(
        deposit.token_index == private_state_transition.token_index,
        "token_index not match"
    );
    anyhow::ensure!(
        deposit.amount == private_state_transition.amount,
        "amount not match"
    );
    anyhow::ensure!(
        nullifier == private_state_transition.nullifier,
        "nullifier not match"
    );

    let prev_private_commitment = private_state_transition.prev_private_state.commitment();
    let new_private_commitment = private_state_transition.new_private_state.commitment();

    Ok(ReceiveDepositValue {
        pubkey,
        deposit_salt,
        deposit_index,
        deposit: deposit.clone(),
        deposit_merkle_proof: deposit_merkle_proof.clone(),
        public_state: public_state.clone(),
        private_state_transition: private_state_transition.clone(),
        prev_private_commitment,
        new_private_commitment,
    })
}
