use crate::server::{health, prover};
use actix_web::web;

pub fn setup_routes(cfg: &mut web::ServiceConfig) {
    cfg.service((
        health::health_check,
        prover::get_proof,
        prover::get_proofs,
        prover::generate_proof,
    ));
}
