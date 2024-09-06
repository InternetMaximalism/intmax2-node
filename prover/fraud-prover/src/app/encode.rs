use base64::prelude::*;
use plonky2::field::goldilocks_field::GoldilocksField;
use plonky2::plonk::circuit_data::VerifierCircuitData;
use plonky2::plonk::config::PoseidonGoldilocksConfig;
use plonky2::plonk::proof::CompressedProofWithPublicInputs;
use plonky2::plonk::proof::ProofWithPublicInputs;

const D: usize = 2;
type C = PoseidonGoldilocksConfig;
type F = GoldilocksField;

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
