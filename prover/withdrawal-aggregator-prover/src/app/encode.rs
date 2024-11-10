use base64::prelude::*;
use intmax2_zkp::common::witness::transfer_witness::TransferWitness;
use plonky2::field::goldilocks_field::GoldilocksField;
use plonky2::plonk::circuit_data::CommonCircuitData;
use plonky2::plonk::circuit_data::VerifierCircuitData;
use plonky2::plonk::config::PoseidonGoldilocksConfig;
use plonky2::plonk::proof::CompressedProofWithPublicInputs;
use plonky2::plonk::proof::ProofWithPublicInputs;
use plonky2::util::serialization::Buffer;
use plonky2::util::serialization::IoResult;
use plonky2::util::serialization::Read;
use plonky2::util::serialization::Write;
use serde::Deserialize;
use serde::Serialize;

const D: usize = 2;
type C = PoseidonGoldilocksConfig;
type F = GoldilocksField;

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SerializableWithdrawalWitness {
    pub single_withdrawal_proof: String,
}

pub struct SingleWithdrawalWitness<F: GoldilocksField, C: PoseidonGoldilocksConfig, const D: usize>
{
    pub single_withdrawal_proof: ProofWithPublicInputs<F, C, D>,
}

impl SerializableWithdrawalWitness {
    pub fn decode(
        &self,
        withdrawal_circuit_data: &VerifierCircuitData<F, C, D>,
    ) -> anyhow::Result<SingleWithdrawalWitness<F, C, D>> {
        let single_withdrawal_proof =
            decode_plonky2_proof(&self.single_withdrawal_proof, withdrawal_circuit_data)?;
        Ok(SingleWithdrawalWitness {
            single_withdrawal_proof,
        })
    }
}

pub(crate) fn encode_plonky2_proof(
    proof: ProofWithPublicInputs<F, C, D>,
    circuit_data: &VerifierCircuitData<F, C, D>,
) -> anyhow::Result<String> {
    let compressed_proof = proof
        .compress(
            &circuit_data.verifier_only.circuit_digest,
            &circuit_data.common,
        )
        .map_err(|e| anyhow::anyhow!("Failed to compress proof: {}", e))?;

    let proof_bytes = compressed_proof_to_bytes(&compressed_proof)
        .map_err(|e| anyhow::anyhow!("Failed to serialize proof: {}", e))?;

    Ok(BASE64_STANDARD.encode(&proof_bytes))
}

pub(crate) fn decode_plonky2_proof(
    encoded_proof: &str,
    circuit_data: &VerifierCircuitData<F, C, D>,
) -> anyhow::Result<ProofWithPublicInputs<F, C, D>> {
    let decoded_proof = BASE64_STANDARD.decode(&encoded_proof)?;
    let compressed_proof = compressed_proof_from_bytes(decoded_proof, &circuit_data.common)
        .map_err(|e| anyhow::anyhow!(e))?;

    compressed_proof.decompress(
        &circuit_data.verifier_only.circuit_digest,
        &circuit_data.common,
    )
}

pub(crate) fn compressed_proof_to_bytes(
    compressed_proof_with_pis: &CompressedProofWithPublicInputs<F, C, D>,
) -> IoResult<Vec<u8>> {
    let mut buffer = Vec::new();

    let CompressedProofWithPublicInputs {
        proof,
        public_inputs,
    } = compressed_proof_with_pis;

    buffer.write_u32(public_inputs.len() as u32)?;
    buffer.write_field_vec(public_inputs)?;
    buffer.write_compressed_proof(proof)?;

    Ok(buffer)
}

pub(crate) fn compressed_proof_from_bytes(
    bytes: Vec<u8>,
    common_data: &CommonCircuitData<F, D>,
) -> IoResult<CompressedProofWithPublicInputs<F, C, D>> {
    let mut buffer = Buffer::new(&bytes);

    let public_inputs_len = buffer.read_u32()?;
    let public_inputs = buffer.read_field_vec(public_inputs_len as usize)?;
    let proof = buffer.read_compressed_proof(common_data)?;

    Ok(CompressedProofWithPublicInputs {
        proof,
        public_inputs,
    })
}
