use anyhow::Context as _;
use base64::prelude::*;
use intmax2_zkp::{
    circuits::{
        balance::balance_processor::BalanceProcessor, validity::validity_circuit::ValidityCircuit,
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
    encode::{decode_plonky2_proof, decode_plonky2_proof_original, encode_plonky2_proof},
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
        // let validity_proof = decode_plonky2_proof(&self.validity_proof, balance_circuit_data)?;
        let validity_proof =
            decode_plonky2_proof_original(&self.validity_proof, balance_circuit_data)?;
        let encoded_validity_proof =
            encode_plonky2_proof(validity_proof.clone(), balance_circuit_data)?;
        println!("encoded_validity_proof: {}", encoded_validity_proof);
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
        let balance_proof = decode_plonky2_proof(&self.balance_proof, balance_circuit_data)?;
        Ok(ReceiveTransferWitness {
            transfer_witness: self.transfer_witness.clone(),
            private_witness: self.private_witness.clone(),
            balance_proof,
            block_merkle_proof: self.block_merkle_proof.clone(),
        })
    }
}

pub async fn generate_receive_deposit_proof_job(
    request_id: String,
    public_key: U256,
    prev_balance_proof: Option<ProofWithPublicInputs<F, C, D>>,
    receive_deposit_witness: &ReceiveDepositWitness,
    balance_processor: &BalanceProcessor<F, C, D>,
    conn: &mut redis::aio::Connection,
) -> anyhow::Result<()> {
    let balance_circuit_data = balance_processor.balance_circuit.data.verifier_data();

    log::debug!("Proving...");
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

    log::debug!("Proving...");
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

    log::debug!("Proving...");
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

    log::debug!("Proving...");
    let balance_proof = balance_processor.prove_send(
        &validity_circuit,
        public_key,
        &send_witness,
        &balance_update_witness,
        &prev_balance_proof,
    );

    let encoded_compressed_balance_proof =
        encode_plonky2_proof(balance_proof, &balance_circuit_data)?;

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
