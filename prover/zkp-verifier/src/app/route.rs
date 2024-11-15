use crate::server::{health, verifier};
use actix_web::web;

pub fn setup_routes(cfg: &mut web::ServiceConfig) {
    cfg.service((health::health_check,));
    cfg.service((verifier::spend::verify_proof,));
}
