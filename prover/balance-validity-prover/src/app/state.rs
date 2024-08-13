use std::sync::{Arc, OnceLock};

use intmax2_zkp::circuits::{
    balance::balance_processor::BalanceProcessor, validity::validity_processor::ValidityProcessor,
};
use plonky2::{field::goldilocks_field::GoldilocksField, plonk::config::PoseidonGoldilocksConfig};

type C = PoseidonGoldilocksConfig;
const D: usize = 2;
type F = GoldilocksField;

pub struct AppState {
    pub balance_processor: Arc<OnceLock<BalanceProcessor<F, C, D>>>,
}

impl AppState {
    pub fn new() -> Self {
        let balance_processor = Arc::new(OnceLock::new());
        let _: tokio::task::JoinHandle<()> =
            tokio::spawn(build_circuits(Arc::clone(&balance_processor)));

        Self { balance_processor }
    }
}

impl Clone for AppState {
    fn clone(&self) -> Self {
        Self {
            balance_processor: Arc::clone(&self.balance_processor),
        }
    }
}

async fn build_circuits(balance_processor_state: Arc<OnceLock<BalanceProcessor<F, C, D>>>) {
    let validity_processor = ValidityProcessor::new();
    let balance_processor = BalanceProcessor::new(&validity_processor.validity_circuit);
    println!("The balance validity circuit build has been completed.");

    let _ = balance_processor_state.get_or_init(|| balance_processor);
}
