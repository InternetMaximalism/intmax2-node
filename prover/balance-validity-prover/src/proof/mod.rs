use anyhow::Context as _;
use intmax2_zkp::{
    circuits::{
        balance::{
            balance_pis::BalancePublicInputs, balance_processor::BalanceProcessor,
            send::spent_circuit::SpentValue,
        },
        validity::validity_circuit::ValidityCircuit,
    },
    common::{
        trees::{account_tree::AccountMembershipProof, block_hash_tree::BlockHashMerkleProof},
        witness::{
            private_witness::PrivateWitness, receive_deposit_witness::ReceiveDepositWitness,
            receive_transfer_witness::ReceiveTransferWitness, send_witness::SendWitness,
            transfer_witness::TransferWitness, update_witness::UpdateWitness,
        },
    },
    ethereum_types::u256::U256,
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
};

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
    let balance_circuit_data = balance_processor.balance_circuit.data.verifier_data();

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
    let balance_proof = balance_processor.prove_receive_deposit(
        public_key,
        &receive_deposit_witness,
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
) -> anyhow::Result<()> {
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

    log::debug!("Proving...");
    let balance_proof = balance_processor.prove_receive_transfer(
        public_key,
        receive_transfer_witness,
        &prev_balance_proof,
    );

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
    send_witness: &SendWitness,
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

pub fn generate_balance_spent_proof_job(
    send_witness: &SendWitness,
    balance_processor: &BalanceProcessor<F, C, D>,
) -> anyhow::Result<String> {
    let spent_circuit = &balance_processor
        .balance_transition_processor
        .sender_processor
        .spent_circuit;

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
        encode_plonky2_proof(spent_proof, &spent_circuit.data.verifier_data())
            .map_err(|e| anyhow::anyhow!("Failed to encode balance proof: {:?}", e))?;


    Ok(encoded_compressed_spent_proof)
}

pub fn generate_balance_single_send_proof_job(
    send_witness: &SendWitness,
    update_witness: &UpdateWitness<F, C, D>,
    balance_processor: &BalanceProcessor<F, C, D>,
    validity_circuit: &ValidityCircuit<F, C, D>,
) -> anyhow::Result<String> {
    let sender_processor = &balance_processor
        .balance_transition_processor.sender_processor;

    log::debug!("Proving...");
    let send_proof = sender_processor.prove(&validity_circuit, &send_witness, &update_witness);

    let encoded_compressed_send_proof =
        encode_plonky2_proof(send_proof, &sender_processor.sender_circuit.data.verifier_data())
            .map_err(|e| anyhow::anyhow!("Failed to encode balance proof: {:?}", e))?;

    Ok(encoded_compressed_send_proof)
}

