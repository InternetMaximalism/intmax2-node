use base64::prelude::*;
use intmax2_zkp::common::trees::account_tree::AccountMembershipProof;
use intmax2_zkp::common::trees::block_hash_tree::BlockHashMerkleProof;
use intmax2_zkp::common::witness::private_witness::PrivateWitness;
use intmax2_zkp::common::witness::receive_transfer_witness::ReceiveTransferWitness;
use intmax2_zkp::common::witness::transfer_witness::TransferWitness;
use intmax2_zkp::common::witness::update_witness::UpdateWitness;
use plonky2::field::goldilocks_field::GoldilocksField;
use plonky2::plonk::circuit_data::VerifierCircuitData;
use plonky2::plonk::config::PoseidonGoldilocksConfig;
use plonky2::plonk::proof::CompressedProofWithPublicInputs;
use plonky2::plonk::proof::ProofWithPublicInputs;
use serde::Deserialize;
use serde::Serialize;

const D: usize = 2;
type C = PoseidonGoldilocksConfig;
type F = GoldilocksField;

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
        let balance_proof = decode_plonky2_proof(&self.balance_proof, balance_circuit_data)?;
        Ok(ReceiveTransferWitness {
            transfer_witness: self.transfer_witness.clone(),
            private_witness: self.private_witness.clone(),
            balance_proof,
            block_merkle_proof: self.block_merkle_proof.clone(),
        })
    }
}

pub(crate) fn encode_plonky2_proof(
    proof: ProofWithPublicInputs<F, C, D>,
    circuit_data: &VerifierCircuitData<F, C, D>,
) -> String {
    let compressed_proof = proof
        .compress(
            &circuit_data.verifier_only.circuit_digest,
            &circuit_data.common,
        )
        .expect("Failed to compress proof");

    BASE64_STANDARD.encode(&compressed_proof.to_bytes())
}

pub(crate) fn decode_plonky2_proof(
    encoded_proof: &str,
    circuit_data: &VerifierCircuitData<F, C, D>,
) -> anyhow::Result<ProofWithPublicInputs<F, C, D>> {
    let decoded_proof = BASE64_STANDARD.decode(&encoded_proof)?;
    let compressed_proof =
        CompressedProofWithPublicInputs::from_bytes(decoded_proof, &circuit_data.common)?;

    compressed_proof.decompress(
        &circuit_data.verifier_only.circuit_digest,
        &circuit_data.common,
    )
}
