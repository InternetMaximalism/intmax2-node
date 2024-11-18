use std::sync::{Arc, OnceLock};

use intmax2_zkp::circuits::{
    balance::balance_processor::BalanceProcessor,
    validity::{validity_circuit::ValidityCircuit, validity_processor::ValidityProcessor},
    withdrawal::withdrawal_processor::WithdrawalProcessor,
};
use plonky2::{field::goldilocks_field::GoldilocksField, plonk::config::PoseidonGoldilocksConfig};

type C = PoseidonGoldilocksConfig;
const D: usize = 2;
type F = GoldilocksField;

pub struct AppState {
    pub withdrawal_processor: Arc<OnceLock<WithdrawalProcessor<F, C, D>>>,
    pub balance_processor: Arc<OnceLock<BalanceProcessor<F, C, D>>>,
    pub validity_circuit: Arc<OnceLock<ValidityCircuit<F, C, D>>>,
}

impl AppState {
    pub fn new() -> Self {
        let withdrawal_processor = Arc::new(OnceLock::new());
        let balance_processor = Arc::new(OnceLock::new());
        let validity_circuit = Arc::new(OnceLock::new());
        let _: tokio::task::JoinHandle<()> = tokio::spawn(build_circuits(
            Arc::clone(&withdrawal_processor),
            Arc::clone(&balance_processor),
            Arc::clone(&validity_circuit),
        ));

        Self {
            withdrawal_processor,
            balance_processor,
            validity_circuit,
        }
    }
}

impl Clone for AppState {
    fn clone(&self) -> Self {
        Self {
            withdrawal_processor: Arc::clone(&self.withdrawal_processor),
            balance_processor: Arc::clone(&self.balance_processor),
            validity_circuit: Arc::clone(&self.validity_circuit),
        }
    }
}

async fn build_circuits(
    withdrawal_processor_state: Arc<OnceLock<WithdrawalProcessor<F, C, D>>>,
    balance_processor_state: Arc<OnceLock<BalanceProcessor<F, C, D>>>,
    validity_circuit_state: Arc<OnceLock<ValidityCircuit<F, C, D>>>,
) {
    let validity_processor = ValidityProcessor::new();
    let balance_processor = BalanceProcessor::new(&validity_processor.validity_circuit);
    let withdrawal_processor =
        WithdrawalProcessor::new(&balance_processor.get_verifier_data().common);
    let _ = validity_circuit_state.get_or_init(|| validity_processor.validity_circuit);
    let _ = balance_processor_state.get_or_init(|| balance_processor);
    let _ = withdrawal_processor_state.get_or_init(|| withdrawal_processor);
    log::info!("The balance validity circuit build has been completed.");
}
