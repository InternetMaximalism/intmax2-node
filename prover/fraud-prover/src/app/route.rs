use crate::server::{health, prover};
use actix_web::web;

pub fn setup_routes(cfg: &mut web::ServiceConfig) {
    cfg.service((health::health_check,));
    cfg.service((
        prover::fraud::get_proof,
        prover::fraud::get_proofs,
        prover::fraud::generate_proof,
    ));
}
