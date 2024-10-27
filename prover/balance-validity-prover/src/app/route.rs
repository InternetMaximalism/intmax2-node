use crate::server::{health, prover};
use actix_web::web;

pub fn setup_routes(cfg: &mut web::ServiceConfig) {
    cfg.service((health::health_check,));
    cfg.service((
        prover::deposit::get_proof,
        prover::deposit::get_proofs,
        prover::deposit::generate_proof,
    ));
    cfg.service((
        prover::update::get_proof,
        prover::update::get_proofs,
        prover::update::generate_proof,
    ));
    cfg.service((
        prover::transfer::get_proof,
        prover::transfer::get_proofs,
        prover::transfer::generate_proof,
    ));
    cfg.service((
        prover::send::get_proof,
        prover::send::get_proofs,
        prover::send::generate_proof,
    ));
    cfg.service((
        prover::spend::get_proof,
        prover::spend::get_proofs,
        prover::spend::generate_proof,
    ));
}
