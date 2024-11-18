use base64::prelude::*;
use plonky2::field::goldilocks_field::GoldilocksField;
use plonky2::plonk::circuit_data::CommonCircuitData;
use plonky2::plonk::circuit_data::VerifierCircuitData;
use plonky2::plonk::config::PoseidonGoldilocksConfig;
use plonky2::plonk::proof::CompressedProofWithPublicInputs;
use plonky2::plonk::proof::ProofWithPublicInputs;
use plonky2::util::serialization::Buffer;
use plonky2::util::serialization::IoResult;
use plonky2::util::serialization::Read;

const D: usize = 2;
type C = PoseidonGoldilocksConfig;
type F = GoldilocksField;

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
