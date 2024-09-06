use std::sync::{Arc, OnceLock};

use intmax2_zkp::circuits::{
    balance::{balance_circuit::BalanceCircuit, balance_processor::BalanceProcessor},
    validity::validity_processor::ValidityProcessor,
    withdrawal::{withdrawal_processor::WithdrawalProcessor, withdrawal_wrapper_processor::WithdrawalWrapperProcessor},
};
use plonky2::{field::goldilocks_field::GoldilocksField, plonk::config::PoseidonGoldilocksConfig};

type C = PoseidonGoldilocksConfig;
const D: usize = 2;
type F = GoldilocksField;

pub struct AppState {
    pub withdrawal_wrapper_processor: Arc<OnceLock<WithdrawalWrapperProcessor>>,
    pub withdrawal_processor: Arc<OnceLock<WithdrawalProcessor<F, C, D>>>,
    pub balance_circuit: Arc<OnceLock<BalanceCircuit<F, C, D>>>,
}

impl AppState {
    pub fn new() -> Self {
        let withdrawal_wrapper_processor = Arc::new(OnceLock::new());
        let withdrawal_processor = Arc::new(OnceLock::new());
        let balance_circuit = Arc::new(OnceLock::new());
        let _: tokio::task::JoinHandle<()> = tokio::spawn(build_circuits(
            Arc::clone(&withdrawal_wrapper_processor),
            Arc::clone(&withdrawal_processor),
            Arc::clone(&balance_circuit),
        ));

        Self {
            withdrawal_wrapper_processor,
            withdrawal_processor,
            balance_circuit,
        }
    }
}

impl Clone for AppState {
    fn clone(&self) -> Self {
        Self {
            withdrawal_wrapper_processor: Arc::clone(&self.withdrawal_wrapper_processor),
            withdrawal_processor: Arc::clone(&self.withdrawal_processor),
            balance_circuit: Arc::clone(&self.balance_circuit),
        }
    }
}

async fn build_circuits(
    withdrawal_wrapper_processor_state: Arc<OnceLock<WithdrawalWrapperProcessor>>,
    withdrawal_processor_state: Arc<OnceLock<WithdrawalProcessor<F, C, D>>>,
    balance_circuit_state: Arc<OnceLock<BalanceCircuit<F, C, D>>>,
) {
    let validity_processor = ValidityProcessor::new();
    let balance_processor = BalanceProcessor::new(&validity_processor.validity_circuit);
    let withdrawal_processor = WithdrawalProcessor::new(&balance_processor.balance_circuit);
    let withdrawal_wrapper_processor = WithdrawalWrapperProcessor::new(&withdrawal_processor.withdrawal_circuit);
    log::info!("The balance validity circuit build has been completed.");

    let _ = withdrawal_wrapper_processor_state.get_or_init(|| withdrawal_wrapper_processor);
    let _ = withdrawal_processor_state.get_or_init(|| withdrawal_processor);
    let _ = balance_circuit_state.get_or_init(|| balance_processor.balance_circuit);
}
