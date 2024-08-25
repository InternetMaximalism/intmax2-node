use std::sync::{Arc, OnceLock};

use intmax2_zkp::circuits::{
    fraud::fraud_processor::FraudProcessor,
    validity::{validity_circuit::ValidityCircuit, validity_processor::ValidityProcessor},
};
use plonky2::{field::goldilocks_field::GoldilocksField, plonk::config::PoseidonGoldilocksConfig};

type C = PoseidonGoldilocksConfig;
const D: usize = 2;
type F = GoldilocksField;

pub struct AppState {
    pub fraud_processor: Arc<OnceLock<FraudProcessor>>,
    pub validity_circuit: Arc<OnceLock<ValidityCircuit<F, C, D>>>,
}

impl AppState {
    pub fn new() -> Self {
        let fraud_processor = Arc::new(OnceLock::new());
        let validity_circuit = Arc::new(OnceLock::new());
        let _: tokio::task::JoinHandle<()> = tokio::spawn(build_circuits(
            Arc::clone(&fraud_processor),
            Arc::clone(&validity_circuit),
        ));

        Self {
            fraud_processor,
            validity_circuit,
        }
    }
}

impl Clone for AppState {
    fn clone(&self) -> Self {
        Self {
            fraud_processor: Arc::clone(&self.fraud_processor),
            validity_circuit: Arc::clone(&self.validity_circuit),
        }
    }
}

async fn build_circuits(
    fraud_processor_state: Arc<OnceLock<FraudProcessor>>,
    validity_circuit_state: Arc<OnceLock<ValidityCircuit<F, C, D>>>,
) {
    let validity_processor = ValidityProcessor::new();
    let fraud_processor = FraudProcessor::new(&validity_processor.validity_circuit);
    log::info!("The fraud circuit build has been completed.");

    let _ = fraud_processor_state.get_or_init(|| fraud_processor);
    let _ = validity_circuit_state.get_or_init(|| validity_processor.validity_circuit);
}
