use crate::app::{
    encode::decode_plonky2_proof,
    interface::{VerifyProofRequest, VerifyProofResponse},
    state::AppState,
};
use actix_web::{post, web, HttpResponse, Responder, Result};

#[post("/verify/spend")]
pub async fn verify_proof(
    req: web::Json<VerifyProofRequest>,
    state: web::Data<AppState>,
) -> Result<impl Responder> {
    let instant = std::time::Instant::now();

    let balance_processor = state
        .balance_processor
        .get()
        .expect("balance processor not initialized");
    let spent_circuit = &balance_processor
        .balance_transition_processor
        .sender_processor
        .spent_circuit;

    let decoded_spend_proof =
        decode_plonky2_proof(&req.proof, &spent_circuit.data.verifier_data()).unwrap();

    log::debug!("Verify...");
    let response = spent_circuit.data.verify(decoded_spend_proof);

    log::info!("Verify completed");
    let response = VerifyProofResponse {
        success: response.is_ok(),
    };
    println!("Verifying time: {:?}", instant.elapsed());

    Ok(HttpResponse::Ok().json(response))
}
