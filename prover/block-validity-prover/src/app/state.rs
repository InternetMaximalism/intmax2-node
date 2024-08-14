use std::sync::{Arc, OnceLock};

use intmax2_zkp::circuits::validity::validity_processor::ValidityProcessor;
use plonky2::{field::goldilocks_field::GoldilocksField, plonk::config::PoseidonGoldilocksConfig};

type C = PoseidonGoldilocksConfig;
const D: usize = 2;
type F = GoldilocksField;

pub struct AppState {
    pub validity_processor: Arc<OnceLock<ValidityProcessor<F, C, D>>>,
}

impl AppState {
    pub fn new() -> Self {
        let validity_processor = Arc::new(OnceLock::new());
        let _: tokio::task::JoinHandle<()> =
            tokio::spawn(build_circuits(Arc::clone(&validity_processor)));

        Self { validity_processor }
    }
}

impl Clone for AppState {
    fn clone(&self) -> Self {
        Self {
            validity_processor: Arc::clone(&self.validity_processor)
        }
    }
}

async fn build_circuits(validity_processor_state: Arc<OnceLock<ValidityProcessor<F, C, D>>>) {
    let validity_processor = ValidityProcessor::new();
    log::info!("The validity circuit build has been completed.");

    let _ = validity_processor_state.get_or_init(|| validity_processor);
}
