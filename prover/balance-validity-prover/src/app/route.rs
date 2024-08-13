use crate::server::{health, prover};
use actix_web::web;

pub fn setup_routes(cfg: &mut web::ServiceConfig) {
    cfg.service((
        health::health_check,
        prover::deposit::get_proof,
        prover::deposit::get_proofs,
        prover::deposit::generate_proof,
    ));
}
