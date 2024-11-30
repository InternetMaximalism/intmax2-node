use std::{
    io::{Read as _, Write as _},
    sync::{Arc, OnceLock},
    time,
};

use intmax2_zkp::circuits::{
    balance::{
        receive::{
            receive_deposit_circuit::ReceiveDepositCircuit,
            receive_transfer_circuit::ReceiveTransferCircuit,
        },
        send::spent_circuit::{SpentCircuit, SpentTarget},
    },
    validity::validity_processor::ValidityProcessor,
};
use plonky2::{
    field::goldilocks_field::GoldilocksField,
    plonk::{
        circuit_data::{CommonCircuitData, VerifierCircuitData},
        config::PoseidonGoldilocksConfig,
    },
    util::serialization::{
        Buffer, GateSerializer, Read as _, WitnessGeneratorSerializer, Write as _,
    },
};

use crate::proof::serializer::{ExtendedGateSerializer, ExtendedGeneratorSerializer};

type C = PoseidonGoldilocksConfig;
const D: usize = 2;
type F = GoldilocksField;

pub struct AppState {
    pub spent_circuit: Arc<OnceLock<SpentCircuit<F, C, D>>>,
    pub receive_transfer_circuit: Arc<OnceLock<ReceiveTransferCircuit<F, C, D>>>,
    pub receive_deposit_circuit: Arc<OnceLock<ReceiveDepositCircuit<F, C, D>>>,
    pub validity_verifier_data: Arc<OnceLock<VerifierCircuitData<F, C, D>>>,
    pub balance_verifier_data: Arc<OnceLock<VerifierCircuitData<F, C, D>>>,
}

impl AppState {
    pub fn new() -> Self {
        let spent_circuit = Arc::new(OnceLock::new());
        let receive_transfer_circuit = Arc::new(OnceLock::new());
        let receive_deposit_circuit = Arc::new(OnceLock::new());
        let validity_verifier_data = Arc::new(OnceLock::new());
        let balance_verifier_data = Arc::new(OnceLock::new());
        let _: tokio::task::JoinHandle<()> = tokio::spawn(build_circuits_async(
            Arc::clone(&spent_circuit),
            Arc::clone(&receive_transfer_circuit),
            Arc::clone(&receive_deposit_circuit),
            Arc::clone(&validity_verifier_data),
            Arc::clone(&balance_verifier_data),
        ));

        Self {
            spent_circuit,
            receive_transfer_circuit,
            receive_deposit_circuit,
            validity_verifier_data,
            balance_verifier_data,
        }
    }
}

impl Clone for AppState {
    fn clone(&self) -> Self {
        Self {
            spent_circuit: Arc::clone(&self.spent_circuit),
            receive_transfer_circuit: Arc::clone(&self.receive_transfer_circuit),
            receive_deposit_circuit: Arc::clone(&self.receive_deposit_circuit),
            validity_verifier_data: Arc::clone(&self.validity_verifier_data),
            balance_verifier_data: Arc::clone(&self.balance_verifier_data),
        }
    }
}

async fn build_circuits_async(
    spent_circuit_state: Arc<OnceLock<SpentCircuit<F, C, D>>>,
    receive_transfer_circuit_state: Arc<OnceLock<ReceiveTransferCircuit<F, C, D>>>,
    receive_deposit_circuit_state: Arc<OnceLock<ReceiveDepositCircuit<F, C, D>>>,
    validity_verifier_data_state: Arc<OnceLock<VerifierCircuitData<F, C, D>>>,
    balance_verifier_data_state: Arc<OnceLock<VerifierCircuitData<F, C, D>>>,
) {
    build_circuits(
        spent_circuit_state,
        receive_transfer_circuit_state,
        receive_deposit_circuit_state,
        validity_verifier_data_state,
        balance_verifier_data_state,
    )
}

fn build_circuits(
    spent_circuit_state: Arc<OnceLock<SpentCircuit<F, C, D>>>,
    receive_transfer_circuit_state: Arc<OnceLock<ReceiveTransferCircuit<F, C, D>>>,
    receive_deposit_circuit_state: Arc<OnceLock<ReceiveDepositCircuit<F, C, D>>>,
    validity_verifier_data_state: Arc<OnceLock<VerifierCircuitData<F, C, D>>>,
    balance_verifier_data_state: Arc<OnceLock<VerifierCircuitData<F, C, D>>>,
) {
    let gate_serializer = ExtendedGateSerializer;
    let generator_serializer = ExtendedGeneratorSerializer::<C, D>::default();

    let build_start = time::Instant::now();
    let validity_processor = ValidityProcessor::<F, C, D>::new();
    println!(
        "validity_processor Building time: {:?}",
        build_start.elapsed()
    );

    {
        let validity_verifier_data = validity_processor.validity_circuit.data.verifier_data();

        let mut serialized_validity_verifier_data: Vec<u8> = vec![];
        serialized_validity_verifier_data
            .write_verifier_circuit_data(&validity_verifier_data, &gate_serializer)
            .unwrap();
        let mut file =
            std::fs::File::create("cache/serialized_validity_verifier_data.txt").unwrap();
        file.write_all(serialized_validity_verifier_data.as_slice())
            .unwrap();
    }

    let start = time::Instant::now();
    let validity_verifier_data = {
        let mut file = std::fs::File::open("cache/serialized_validity_verifier_data.txt").unwrap();
        let mut serialized_validity_verifier_data: Vec<u8> = vec![];
        file.read_to_end(&mut serialized_validity_verifier_data)
            .unwrap();
        println!(
            "size of serialized_validity_verifier_data: {}",
            serialized_validity_verifier_data.len()
        );

        let mut reader = Buffer::new(&serialized_validity_verifier_data);

        reader.read_verifier_circuit_data(&gate_serializer).unwrap()
    };
    println!("Decoding time: {:?}", start.elapsed());
    println!("Complete validity_verifier_data");

    // let balance_transition_processor =
    //     BalanceTransitionProcessor::<F, C, D>::new(&validity_processor.validity_circuit);
    // let balance_circuit =
    //     BalanceCircuit::new(&balance_transition_processor.balance_transition_circuit);

    // {
    //     let balance_verifier_data = balance_circuit.data.verifier_data();

    //     let mut serialized_balance_verifier_data: Vec<u8> = vec![];
    //     serialized_balance_verifier_data
    //         .write_verifier_circuit_data(&balance_verifier_data, &gate_serializer)
    //         .unwrap();
    //     let mut file = std::fs::File::create("cache/serialized_balance_verifier_data.txt").unwrap();
    //     file.write_all(serialized_balance_verifier_data.as_slice())
    //         .unwrap();
    // }

    let start = time::Instant::now();
    let balance_verifier_data = {
        let mut file = std::fs::File::open("cache/serialized_balance_verifier_data.txt").unwrap();
        let mut serialized_balance_verifier_data: Vec<u8> = vec![];
        file.read_to_end(&mut serialized_balance_verifier_data)
            .unwrap();
        println!(
            "size of serialized_balance_verifier_data: {}",
            serialized_balance_verifier_data.len()
        );

        let mut reader = Buffer::new(&serialized_balance_verifier_data);

        reader.read_verifier_circuit_data(&gate_serializer).unwrap()
    };
    println!("Decoding time: {:?}", start.elapsed());
    println!("Complete balance_verifier_data");

    build_spent_circuit(spent_circuit_state, &gate_serializer, &generator_serializer);

    build_receive_transfer_circuit(
        receive_transfer_circuit_state,
        &gate_serializer,
        &generator_serializer,
        &balance_verifier_data.common,
    );

    build_receive_deposit_circuit(
        receive_deposit_circuit_state,
        &gate_serializer,
        &generator_serializer,
    );

    let _ = validity_verifier_data_state.get_or_init(|| validity_verifier_data);
    let _ = balance_verifier_data_state.get_or_init(|| balance_verifier_data);
}

fn build_spent_circuit(
    spent_circuit_state: Arc<OnceLock<SpentCircuit<F, C, D>>>,
    gate_serializer: &dyn GateSerializer<F, D>,
    generator_serializer: &dyn WitnessGeneratorSerializer<F, D>,
) {
    let start = time::Instant::now();
    let decoded_spent_circuit = {
        let mut file = std::fs::File::open("cache/serialized_spent_circuit_data.txt").unwrap();
        let mut serialized_spent_circuit_data: Vec<u8> = vec![];
        file.read_to_end(&mut serialized_spent_circuit_data)
            .unwrap();
        println!(
            "size of serialized_spent_circuit_data: {}",
            serialized_spent_circuit_data.len()
        );
        let mut reader = Buffer::new(&serialized_spent_circuit_data);
        let decoded_spent_circuit_data = reader
            .read_circuit_data::<F, C, D>(gate_serializer, generator_serializer)
            .unwrap();
        let spent_circuit_target = SpentTarget::from_buffer(&mut reader).unwrap();
        SpentCircuit {
            target: spent_circuit_target,
            data: decoded_spent_circuit_data,
        }
    };
    println!("Decoding time: {:?}", start.elapsed());
    println!("Complete spent_circuit");

    let _ = spent_circuit_state.get_or_init(|| decoded_spent_circuit);
}

fn build_receive_deposit_circuit(
    receive_deposit_circuit_state: Arc<OnceLock<ReceiveDepositCircuit<F, C, D>>>,
    gate_serializer: &dyn GateSerializer<F, D>,
    generator_serializer: &dyn WitnessGeneratorSerializer<F, D>,
) {
    // let build_start = time::Instant::now();
    // let receive_deposit_circuit = ReceiveDepositCircuit::<F, C, D>::new();
    // println!(
    //     "receive deposit circuit degree bits: {}",
    //     receive_deposit_circuit.data.common.degree_bits()
    // );
    // {
    //     let mut serialized_receive_deposit_circuit_data: Vec<u8> = vec![];
    //     serialized_receive_deposit_circuit_data
    //         .write_circuit_data(
    //             &receive_deposit_circuit.data,
    //             gate_serializer,
    //             generator_serializer,
    //         )
    //         .unwrap();

    //     receive_deposit_circuit
    //         .target
    //         .to_buffer(&mut serialized_receive_deposit_circuit_data)
    //         .unwrap();
    //     let mut file =
    //         std::fs::File::create("cache/serialized_receive_deposit_circuit_data.txt").unwrap();
    //     file.write_all(serialized_receive_deposit_circuit_data.as_slice())
    //         .unwrap();
    // }
    // println!("Building time: {:?}", build_start.elapsed());

    let start = time::Instant::now();
    let decoded_receive_deposit_circuit = {
        let mut serialized_receive_deposit_circuit_data: Vec<u8> = vec![];
        let mut file =
            std::fs::File::open("cache/serialized_receive_deposit_circuit_data.txt").unwrap();
        file.read_to_end(&mut serialized_receive_deposit_circuit_data)
            .unwrap();
        println!(
            "size of serialized_receive_deposit_circuit_data: {}",
            serialized_receive_deposit_circuit_data.len()
        );

        let mut reader = Buffer::new(&serialized_receive_deposit_circuit_data);
        ReceiveDepositCircuit::<F, C, D>::from_buffer(
            &mut reader,
            gate_serializer,
            generator_serializer,
        )
        .expect("Failed to decode receive deposit circuit")
    };
    println!("Decoding time: {:?}", start.elapsed());
    println!("Complete receive_deposit_circuit");

    let _ = receive_deposit_circuit_state.get_or_init(|| decoded_receive_deposit_circuit);
}

fn build_receive_transfer_circuit(
    receive_transfer_circuit_state: Arc<OnceLock<ReceiveTransferCircuit<F, C, D>>>,
    gate_serializer: &dyn GateSerializer<F, D>,
    generator_serializer: &dyn WitnessGeneratorSerializer<F, D>,
    _balance_circuit_common_data: &CommonCircuitData<F, D>,
) {
    // let build_start = time::Instant::now();
    // {
    //     let receive_transfer_circuit =
    //         ReceiveTransferCircuit::<F, C, D>::new(&_balance_circuit_common_data);
    //     let receive_transfer_circuit_data = receive_transfer_circuit.data.verifier_data();
    //     println!(
    //         "receive transfer circuit degree bits: {}",
    //         receive_transfer_circuit_data.common.degree_bits()
    //     );

    //     let mut serialized_receive_transfer_circuit_data: Vec<u8> = vec![];
    //     serialized_receive_transfer_circuit_data
    //         .write_circuit_data(
    //             &receive_transfer_circuit.data,
    //             gate_serializer,
    //             generator_serializer,
    //         )
    //         .unwrap();
    //     receive_transfer_circuit
    //         .target
    //         .to_buffer(&mut serialized_receive_transfer_circuit_data)
    //         .unwrap();
    //     let mut file =
    //         std::fs::File::create("cache/serialized_receive_transfer_circuit_data.txt").unwrap();
    //     file.write_all(serialized_receive_transfer_circuit_data.as_slice())
    //         .unwrap();
    // }
    // println!("Building time: {:?}", build_start.elapsed());

    let start = time::Instant::now();
    let decoded_receive_transfer_circuit = {
        let mut file =
            std::fs::File::open("cache/serialized_receive_transfer_circuit_data.txt").unwrap();
        let mut serialized_receive_transfer_circuit_data: Vec<u8> = vec![];
        file.read_to_end(&mut serialized_receive_transfer_circuit_data)
            .unwrap();
        println!(
            "size of serialized_receive_transfer_circuit_data: {}",
            serialized_receive_transfer_circuit_data.len()
        );

        let mut reader = Buffer::new(&serialized_receive_transfer_circuit_data);
        ReceiveTransferCircuit::<F, C, D>::from_buffer(
            &mut reader,
            gate_serializer,
            generator_serializer,
        )
        .unwrap()
    };

    println!("Decoding time: {:?}", start.elapsed());
    println!("Complete receive_transfer_circuit");

    let _ = receive_transfer_circuit_state.get_or_init(|| decoded_receive_transfer_circuit);
}
