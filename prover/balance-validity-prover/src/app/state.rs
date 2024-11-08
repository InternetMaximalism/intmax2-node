use std::sync::{Arc, OnceLock};

use intmax2_zkp::circuits::{
    balance::balance_processor::BalanceProcessor,
    validity::{validity_circuit::ValidityCircuit, validity_processor::ValidityProcessor},
    withdrawal::single_withdrawal_circuit::SingleWithdrawalCircuit,
};
use plonky2::{field::goldilocks_field::GoldilocksField, plonk::config::PoseidonGoldilocksConfig};

type C = PoseidonGoldilocksConfig;
const D: usize = 2;
type F = GoldilocksField;

pub struct AppState {
    pub balance_processor: Arc<OnceLock<BalanceProcessor<F, C, D>>>,
    pub validity_circuit: Arc<OnceLock<ValidityCircuit<F, C, D>>>,
    pub single_withdrawal_circuit: Arc<OnceLock<SingleWithdrawalCircuit<F, C, D>>>,
}

impl AppState {
    pub fn new() -> Self {
        let balance_processor = Arc::new(OnceLock::new());
        let validity_circuit = Arc::new(OnceLock::new());
        let single_withdrawal_circuit = Arc::new(OnceLock::new());
        let _: tokio::task::JoinHandle<()> = tokio::spawn(build_circuits(
            Arc::clone(&balance_processor),
            Arc::clone(&validity_circuit),
            Arc::clone(&single_withdrawal_circuit),
        ));

        Self {
            balance_processor,
            validity_circuit,
            single_withdrawal_circuit,
        }
    }
}

impl Clone for AppState {
    fn clone(&self) -> Self {
        Self {
            balance_processor: Arc::clone(&self.balance_processor),
            validity_circuit: Arc::clone(&self.validity_circuit),
            single_withdrawal_circuit: Arc::clone(&self.single_withdrawal_circuit),
        }
    }
}

async fn build_circuits(
    balance_processor_state: Arc<OnceLock<BalanceProcessor<F, C, D>>>,
    validity_circuit_state: Arc<OnceLock<ValidityCircuit<F, C, D>>>,
    single_withdrawal_circuit_state: Arc<OnceLock<SingleWithdrawalCircuit<F, C, D>>>,
) {
    let validity_processor = ValidityProcessor::new();
    let balance_processor = BalanceProcessor::new(&validity_processor.validity_circuit);
    let single_withdrawal_circuit =
        SingleWithdrawalCircuit::new(&balance_processor.balance_circuit.data.common);
    log::info!("The balance validity circuit build has been completed.");

    let _ = balance_processor_state.get_or_init(|| balance_processor);
    let _ = validity_circuit_state.get_or_init(|| validity_processor.validity_circuit);
    let _ = single_withdrawal_circuit_state.get_or_init(|| single_withdrawal_circuit);
}
