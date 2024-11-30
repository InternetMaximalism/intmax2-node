use crate::server::{health, prover};
use actix_web::web;

pub fn setup_routes(cfg: &mut web::ServiceConfig) {
    cfg.service((health::health_check,));
    cfg.service((
        prover::deposit_transition::get_proof,
        prover::deposit_transition::get_proofs,
        prover::deposit_transition::generate_proof,
    ));
    cfg.service((
        prover::transfer_transition::get_proof,
        prover::transfer_transition::get_proofs,
        prover::transfer_transition::generate_proof,
    ));
    cfg.service((
        prover::send_transition::get_proof,
        prover::send_transition::get_proofs,
        prover::send_transition::generate_proof,
    ));
}
