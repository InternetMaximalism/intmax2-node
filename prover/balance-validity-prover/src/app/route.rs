use crate::server::{health, prover};
use actix_web::web;

pub fn setup_routes(cfg: &mut web::ServiceConfig) {
    cfg.service((
        health::health_check,
        prover::deposit::get_proof,
        prover::deposit::get_proofs,
        prover::deposit::generate_proof,
        prover::update::get_proof,
        prover::update::get_proofs,
        prover::update::generate_proof,
        prover::transfer::get_proof,
        prover::transfer::get_proofs,
        prover::transfer::generate_proof,
    ));
}
