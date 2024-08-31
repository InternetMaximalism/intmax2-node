use base64::prelude::*;
use intmax2_zkp::common::witness::transfer_witness::TransferWitness;
use intmax2_zkp::common::witness::withdrawal_witness::WithdrawalWitness;
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
pub struct SerializableWithdrawalWitness {
    pub transfer_witness: TransferWitness,
    pub balance_proof: String,
}

impl SerializableWithdrawalWitness {
    pub fn decode(
        &self,
        balance_circuit_data: &VerifierCircuitData<F, C, D>,
    ) -> anyhow::Result<WithdrawalWitness<F, C, D>> {
        let balance_proof = decode_plonky2_proof(&self.balance_proof, balance_circuit_data)?;
        Ok(WithdrawalWitness {
            balance_proof,
            transfer_witness: self.transfer_witness.clone(),
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
