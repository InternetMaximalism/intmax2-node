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
            private_witness::PrivateWitness, receive_deposit_witness::ReceiveDepositWitness,
            receive_transfer_witness::ReceiveTransferWitness, send_witness::SendWitness,
            transfer_witness::TransferWitness, update_witness::UpdateWitness,
        },
    },
    ethereum_types::{bytes32::Bytes32, u256::U256, u32limb_trait::U32LimbTrait},
    utils::leafable::Leafable,
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
    interface::SpentTokenWitness,
};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RedisResponse {
    pub success: bool,
    pub message: String,
}

const D: usize = 2;
type C = PoseidonGoldilocksConfig;
type F = <C as GenericConfig<D>>::F;

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SerializableUpdateWitness {
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
    pub private_witness: PrivateWitness,
    pub balance_proof: String,
    pub block_merkle_proof: BlockHashMerkleProof,
}

impl SerializableReceiveTransferWitness {
    pub fn decode(
        &self,
        balance_circuit_data: &VerifierCircuitData<F, C, D>,
    ) -> anyhow::Result<ReceiveTransferWitness<F, C, D>> {
        let balance_proof = decode_plonky2_proof(&self.balance_proof, balance_circuit_data)
            .map_err(|e| anyhow::anyhow!("Failed to decode balance proof: {:?}", e))?;
        Ok(ReceiveTransferWitness {
            transfer_witness: self.transfer_witness.clone(),
            private_witness: self.private_witness.clone(),
            balance_proof,
            block_merkle_proof: self.block_merkle_proof.clone(),
        })
    }
}

pub async fn generate_receive_deposit_proof_job(
    full_request_id: String,
    public_key: U256,
    prev_balance_proof: Option<ProofWithPublicInputs<F, C, D>>,
    receive_deposit_witness: &ReceiveDepositWitness,
    balance_processor: &BalanceProcessor<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let result = generate_receive_deposit_proof(
        public_key,
        prev_balance_proof,
        receive_deposit_witness,
        balance_processor,
    );

    match result {
        Ok(encoded_compressed_validity_proof) => {
            let result = RedisResponse {
                success: true,
                message: encoded_compressed_validity_proof,
            };
            let result_json = serde_json::to_string(&result)
                .expect("Failed to serialize success response to JSON");
            let opts = SetOptions::default()
                .conditional_set(ExistenceCheck::NX)
                .get(true)
                .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

            let _ = redis::Cmd::set_options(&full_request_id, result_json, opts)
                .query_async::<_, Option<String>>(conn)
                .await
                .with_context(|| "Failed to set proof")?;

            Ok(())
        }
        Err(e) => {
            log::error!("Failed to generate proof: {:?}", e);
            let result = RedisResponse {
                success: false,
                message: format!("Failed to generate proof: {:?}", e),
            };
            let result_json = serde_json::to_string(&result)
                .expect("Failed to serialize failure response to JSON");
            let opts = SetOptions::default()
                .conditional_set(ExistenceCheck::NX)
                .get(true)
                .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

            let _ = redis::Cmd::set_options(&full_request_id, result_json, opts)
                .query_async::<_, Option<String>>(conn)
                .await
                .with_context(|| "Failed to set proof")?;

            Err(e)
        }
    }
}

pub fn generate_receive_deposit_proof(
    public_key: U256,
    prev_balance_proof: Option<ProofWithPublicInputs<F, C, D>>,
    receive_deposit_witness: &ReceiveDepositWitness,
    balance_processor: &BalanceProcessor<F, C, D>,
) -> anyhow::Result<String> {
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
    let balance_proof = balance_processor.prove_receive_deposit(
        public_key,
        &receive_deposit_witness,
        &prev_balance_proof,
    );

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

    Ok(encoded_compressed_balance_proof)
}

pub async fn generate_balance_update_proof_job(
    full_request_id: String,
    public_key: U256,
    prev_balance_proof: Option<ProofWithPublicInputs<F, C, D>>,
    balance_update_witness: &UpdateWitness<F, C, D>,
    balance_processor: &BalanceProcessor<F, C, D>,
    validity_circuit: &ValidityCircuit<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let result = generate_balance_update_proof(
        public_key,
        prev_balance_proof,
        balance_update_witness,
        balance_processor,
        validity_circuit,
    );

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

            let _ = redis::Cmd::set_options(&full_request_id, response_json, opts)
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

            let _ = redis::Cmd::set_options(&full_request_id, response_json, opts)
                .query_async::<_, Option<String>>(conn)
                .await
                .with_context(|| "Failed to set proof")?;

            Err(e)
        }
    }
}

pub fn generate_balance_update_proof(
    public_key: U256,
    prev_balance_proof: Option<ProofWithPublicInputs<F, C, D>>,
    balance_update_witness: &UpdateWitness<F, C, D>,
    balance_processor: &BalanceProcessor<F, C, D>,
    validity_circuit: &ValidityCircuit<F, C, D>,
) -> anyhow::Result<String> {
    let balance_circuit_data = balance_processor.balance_circuit.data.verifier_data();
    // let validity_circuit_data = validity_circuit_data.verifier_data();

    log::debug!("Proving...");
    let balance_proof = balance_processor.prove_update(
        &validity_circuit,
        public_key,
        &balance_update_witness,
        &prev_balance_proof,
    );

    let encoded_compressed_balance_proof =
        encode_plonky2_proof(balance_proof.clone(), &balance_circuit_data)?;

    Ok(encoded_compressed_balance_proof)
}

pub async fn generate_balance_transfer_proof_job(
    full_request_id: String,
    public_key: U256,
    prev_balance_proof: Option<ProofWithPublicInputs<F, C, D>>,
    receive_transfer_witness: &ReceiveTransferWitness<F, C, D>,
    balance_processor: &BalanceProcessor<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let response = generate_balance_transfer_proof(
        public_key,
        prev_balance_proof,
        receive_transfer_witness,
        balance_processor,
    );

    match response {
        Ok(encoded_compressed_balance_proof) => {
            let response = RedisResponse {
                success: true,
                message: encoded_compressed_balance_proof,
            };
            let response_json = serde_json::to_string(&response)
                .expect("Failed to serialize success response to JSON");
            let opts = SetOptions::default()
                .conditional_set(ExistenceCheck::NX)
                .get(true)
                .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

            let _ = redis::Cmd::set_options(&full_request_id, response_json, opts)
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

            let _ = redis::Cmd::set_options(&full_request_id, response_json, opts)
                .query_async::<_, Option<String>>(conn)
                .await
                .with_context(|| "Failed to set proof")?;

            Err(e)
        }
    }
}

pub fn generate_balance_transfer_proof(
    public_key: U256,
    prev_balance_proof: Option<ProofWithPublicInputs<F, C, D>>,
    receive_transfer_witness: &ReceiveTransferWitness<F, C, D>,
    balance_processor: &BalanceProcessor<F, C, D>,
) -> anyhow::Result<String> {
    let balance_circuit_data = balance_processor.balance_circuit.data.verifier_data();
    // let validity_circuit_data = validity_circuit_data.verifier_data();

    log::debug!("Proving...");
    let balance_proof = balance_processor.prove_receive_transfer(
        public_key,
        receive_transfer_witness,
        &prev_balance_proof,
    );

    let encoded_compressed_balance_proof =
        encode_plonky2_proof(balance_proof.clone(), &balance_circuit_data)?;

    Ok(encoded_compressed_balance_proof)
}

pub async fn generate_balance_send_proof_job(
    full_request_id: String,
    public_key: U256,
    prev_balance_proof: Option<ProofWithPublicInputs<F, C, D>>,
    send_witness: &SendWitness,
    balance_update_witness: &UpdateWitness<F, C, D>,
    balance_processor: &BalanceProcessor<F, C, D>,
    validity_circuit: &ValidityCircuit<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let response = generate_balance_send_proof(
        public_key,
        prev_balance_proof,
        send_witness,
        balance_update_witness,
        balance_processor,
        validity_circuit,
    );

    match response {
        Ok(encoded_compressed_balance_proof) => {
            let response = RedisResponse {
                success: true,
                message: encoded_compressed_balance_proof,
            };
            let response_json = serde_json::to_string(&response)
                .expect("Failed to serialize success response to JSON");
            let opts = SetOptions::default()
                .conditional_set(ExistenceCheck::NX)
                .get(true)
                .with_expiration(SetExpiry::EX(config::get("proof_expiration")));

            let _ = redis::Cmd::set_options(&full_request_id, response_json, opts)
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

            let _ = redis::Cmd::set_options(&full_request_id, response_json, opts)
                .query_async::<_, Option<String>>(conn)
                .await
                .with_context(|| "Failed to set proof")?;

            Err(e)
        }
    }
}

pub fn generate_balance_send_proof(
    public_key: U256,
    prev_balance_proof: Option<ProofWithPublicInputs<F, C, D>>,
    send_witness: &SendWitness,
    balance_update_witness: &UpdateWitness<F, C, D>,
    balance_processor: &BalanceProcessor<F, C, D>,
    validity_circuit: &ValidityCircuit<F, C, D>,
) -> anyhow::Result<String> {
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
    let balance_proof = balance_processor.prove_send(
        &validity_circuit,
        public_key,
        &send_witness,
        &balance_update_witness,
        &prev_balance_proof,
    );

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

    Ok(encoded_compressed_balance_proof)
}

pub fn generate_balance_spend_proof(
    spent_token_witness: &SpentTokenWitness,
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
        spent_token_witness.tx_nonce,
    );

    log::debug!("Proving...");
    let spent_proof = spent_circuit.prove(&spent_value).unwrap();

    let encoded_compressed_spent_proof =
        encode_plonky2_proof(spent_proof, &spent_circuit.data.verifier_data())
            .map_err(|e| anyhow::anyhow!("Failed to encode balance proof: {:?}", e))?;

    Ok(encoded_compressed_spent_proof)
}

pub fn validate_witness(
    _pubkey: U256,
    public_state: &PublicState,
    receive_deposit_witness: &ReceiveDepositWitness,
    prev_balance_proof: &Option<ProofWithPublicInputs<F, C, D>>,
) -> anyhow::Result<()> {
    let deposit_witness = receive_deposit_witness.deposit_witness.clone();
    let private_transition_witness = receive_deposit_witness.private_witness.clone();

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
    );

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
