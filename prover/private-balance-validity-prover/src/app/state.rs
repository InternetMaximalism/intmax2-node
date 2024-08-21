use std::{
    io::{Read as _, Write as _},
    sync::{Arc, OnceLock},
    time,
};

use intmax2_zkp::circuits::{
    balance::{
        balance_processor::BalanceProcessor,
        receive::{
            receive_deposit_circuit::ReceiveDepositCircuit,
            receive_transfer_circuit::ReceiveTransferCircuit,
        },
        send::spent_circuit::{SpentCircuit, SpentTarget},
    },
    validity::{validity_circuit::ValidityCircuit, validity_processor::ValidityProcessor},
};
use plonky2::{
    field::goldilocks_field::GoldilocksField,
    plonk::config::PoseidonGoldilocksConfig,
    util::serialization::{Buffer, Read as _, Write as _},
};

use crate::proof::serializer::{ExtendedGateSerializer, ExtendedGeneratorSerializer};

type C = PoseidonGoldilocksConfig;
const D: usize = 2;
type F = GoldilocksField;

pub struct AppState {
    pub balance_processor: Arc<OnceLock<BalanceProcessor<F, C, D>>>,
    pub validity_circuit: Arc<OnceLock<ValidityCircuit<F, C, D>>>,
}

impl AppState {
    pub fn new() -> Self {
        let balance_processor = Arc::new(OnceLock::new());
        let validity_circuit = Arc::new(OnceLock::new());
        let _: tokio::task::JoinHandle<()> = tokio::spawn(build_circuits(
            Arc::clone(&balance_processor),
            Arc::clone(&validity_circuit),
        ));

        Self {
            balance_processor,
            validity_circuit,
        }
    }
}

impl Clone for AppState {
    fn clone(&self) -> Self {
        Self {
            balance_processor: Arc::clone(&self.balance_processor),
            validity_circuit: Arc::clone(&self.validity_circuit),
        }
    }
}

async fn build_circuits(
    balance_processor_state: Arc<OnceLock<BalanceProcessor<F, C, D>>>,
    validity_circuit_state: Arc<OnceLock<ValidityCircuit<F, C, D>>>,
) {
    let gate_serializer = ExtendedGateSerializer;
    let generator_serializer = ExtendedGeneratorSerializer::<C, D>::default();

    let build_start = time::Instant::now();
    // let spent_circuit = SpentCircuit::<F, C, D>::new();
    // println!(
    //     "spent circuit degree bits: {}",
    //     spent_circuit.data.common.degree_bits()
    // );
    // {
    //     let mut serialized_spent_circuit_data: Vec<u8> = vec![];
    //     serialized_spent_circuit_data
    //         .write_circuit_data(&spent_circuit.data, &gate_serializer, &generator_serializer)
    //         .unwrap();

    //     spent_circuit
    //         .target
    //         .to_buffer(&mut serialized_spent_circuit_data)
    //         .unwrap();
    //     let mut file = std::fs::File::create("serialized_spent_circuit_data.txt").unwrap();
    //     file.write_all(serialized_spent_circuit_data.as_slice())
    //         .unwrap();
    // }

    let start = time::Instant::now();
    let decoded_spent_circuit = {
        let mut file = std::fs::File::open("serialized_spent_circuit_data.txt").unwrap();
        let mut serialized_spent_circuit_data: Vec<u8> = vec![];
        file.read_to_end(&mut serialized_spent_circuit_data)
            .unwrap();
        println!(
            "size of serialized_spent_circuit_data: {}",
            serialized_spent_circuit_data.len()
        );
        let mut reader = Buffer::new(&serialized_spent_circuit_data);
        let decoded_spent_circuit_data = reader
            .read_circuit_data::<F, C, D>(&gate_serializer, &generator_serializer)
            .unwrap();
        let spent_circuit_target = SpentTarget::from_buffer(&mut reader).unwrap();
        let decoded_spent_circuit = SpentCircuit {
            target: spent_circuit_target,
            data: decoded_spent_circuit_data,
        };

        decoded_spent_circuit
    };
    println!("Decoding time: {:?}", start.elapsed());
    println!("Building time: {:?}", build_start.elapsed());

    // if spent_circuit != decoded_spent_circuit {
    //     panic!("mismatched spent_circuit");
    // }
    println!("Complete spent_circuit");

    let build_start = time::Instant::now();
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
    //             &gate_serializer,
    //             &generator_serializer,
    //         )
    //         .unwrap();

    //     receive_deposit_circuit
    //         .target
    //         .to_buffer(&mut serialized_receive_deposit_circuit_data)
    //         .unwrap();
    //     let mut file =
    //         std::fs::File::create("serialized_receive_deposit_circuit_data.txt").unwrap();
    //     file.write_all(serialized_receive_deposit_circuit_data.as_slice())
    //         .unwrap();
    // }

    let start = time::Instant::now();
    let decoded_receive_deposit_circuit = {
        let mut serialized_receive_deposit_circuit_data: Vec<u8> = vec![];
        let mut file = std::fs::File::open("serialized_receive_deposit_circuit_data.txt").unwrap();
        file.read_to_end(&mut serialized_receive_deposit_circuit_data)
            .unwrap();
        println!(
            "size of serialized_receive_deposit_circuit_data: {}",
            serialized_receive_deposit_circuit_data.len()
        );

        let mut reader = Buffer::new(&serialized_receive_deposit_circuit_data);
        let decoded_receive_deposit_circuit = ReceiveDepositCircuit::<F, C, D>::from_buffer(
            &mut reader,
            &gate_serializer,
            &generator_serializer,
        )
        .unwrap();

        decoded_receive_deposit_circuit
    };
    println!("Decoding time: {:?}", start.elapsed());
    println!("Building time: {:?}", build_start.elapsed());

    // if receive_deposit_circuit.data != decoded_receive_deposit_circuit.data {
    //     panic!("mismatched receive_deposit_circuit data");
    // }
    // if receive_deposit_circuit.target != decoded_receive_deposit_circuit.target {
    //     panic!("mismatched receive_deposit_circuit target");
    // }
    println!("Complete receive_deposit_circuit");

    let build_start = time::Instant::now();
    // {
    //     let receive_transfer_circuit_data = balance_processor
    //         .balance_transition_processor
    //         .receive_transfer_circuit
    //         .data
    //         .verifier_data();
    //     println!(
    //         "receive transfer circuit degree bits: {}",
    //         receive_transfer_circuit_data.common.degree_bits()
    //     );

    //     let mut serialized_receive_transfer_circuit_data: Vec<u8> = vec![];
    //     serialized_receive_transfer_circuit_data
    //         .write_circuit_data(
    //             &balance_processor
    //                 .balance_transition_processor
    //                 .receive_transfer_circuit
    //                 .data,
    //             &gate_serializer,
    //             &generator_serializer,
    //         )
    //         .unwrap();
    //     balance_processor
    //         .balance_transition_processor
    //         .receive_transfer_circuit
    //         .target
    //         .to_buffer(&mut serialized_receive_transfer_circuit_data)
    //         .unwrap();
    //     let mut file =
    //         std::fs::File::create("serialized_receive_transfer_circuit_data.txt").unwrap();
    //     file.write_all(serialized_receive_transfer_circuit_data.as_slice())
    //         .unwrap();
    // }

    let start = time::Instant::now();
    let decoded_receive_transfer_circuit = {
        let mut file = std::fs::File::open("serialized_receive_transfer_circuit_data.txt").unwrap();
        let mut serialized_receive_transfer_circuit_data: Vec<u8> = vec![];
        file.read_to_end(&mut serialized_receive_transfer_circuit_data)
            .unwrap();
        println!(
            "size of serialized_receive_transfer_circuit_data: {}",
            serialized_receive_transfer_circuit_data.len()
        );

        let mut reader = Buffer::new(&serialized_receive_transfer_circuit_data);
        let decoded_receive_transfer_circuit = ReceiveTransferCircuit::<F, C, D>::from_buffer(
            &mut reader,
            &gate_serializer,
            &generator_serializer,
        )
        .unwrap();

        decoded_receive_transfer_circuit
    };
    println!("Decoding time: {:?}", start.elapsed());
    println!("Building time: {:?}", build_start.elapsed());

    // if balance_processor
    //     .balance_transition_processor
    //     .receive_transfer_circuit
    //     .data
    //     != decoded_receive_transfer_circuit.data
    // {
    //     panic!("mismatched receive_transfer_circuit data");
    // }
    // if balance_processor
    //     .balance_transition_processor
    //     .receive_transfer_circuit
    //     .target
    //     != decoded_receive_transfer_circuit.target
    // {
    //     panic!("mismatched receive_transfer_circuit target");
    // }
    // println!("Complete receive_transfer_circuit");

    let build_start = time::Instant::now();
    let validity_processor = ValidityProcessor::<F, C, D>::new();
    println!(
        "validity_processor Building time: {:?}",
        build_start.elapsed()
    );
    let build_start = time::Instant::now();
    let balance_processor = BalanceProcessor::new(&validity_processor.validity_circuit);
    log::info!("The balance validity circuit build has been completed.");
    println!(
        "balance_processor Building time: {:?}",
        build_start.elapsed()
    );

    let _ = balance_processor_state.get_or_init(|| balance_processor);
    let _ = validity_circuit_state.get_or_init(|| validity_processor.validity_circuit);
}

#[test]
fn test_receive_deposit_circuit() {
    let gate_serializer = ExtendedGateSerializer;
    let generator_serializer = ExtendedGeneratorSerializer::<C, D>::default();

    let validity_processor = ValidityProcessor::<F, C, D>::new();
    let balance_processor = BalanceProcessor::new(&validity_processor.validity_circuit);
    log::info!("The balance validity circuit build has been completed.");
    {
        let receive_transfer_circuit_data = balance_processor
            .balance_transition_processor
            .receive_transfer_circuit
            .data
            .verifier_data();
        println!(
            "receive transfer circuit degree bits: {}",
            receive_transfer_circuit_data.common.degree_bits()
        );

        let mut serialized_receive_transfer_circuit_data: Vec<u8> = vec![];
        serialized_receive_transfer_circuit_data
            .write_circuit_data(
                &balance_processor
                    .balance_transition_processor
                    .receive_transfer_circuit
                    .data,
                &gate_serializer,
                &generator_serializer,
            )
            .unwrap();
        balance_processor
            .balance_transition_processor
            .receive_transfer_circuit
            .target
            .to_buffer(&mut serialized_receive_transfer_circuit_data)
            .unwrap();

        println!(
            "size of serialized_receive_transfer_circuit_data: {}",
            serialized_receive_transfer_circuit_data.len()
        );
        let mut file =
            std::fs::File::create("serialized_receive_transfer_circuit_data.txt").unwrap();
        file.write_all(serialized_receive_transfer_circuit_data.as_slice())
            .unwrap();
    }

    let decoded_receive_transfer_circuit = {
        let mut file = std::fs::File::open("serialized_receive_transfer_circuit_data.txt").unwrap();
        let mut serialized_receive_transfer_circuit_data: Vec<u8> = vec![];
        file.read_to_end(&mut serialized_receive_transfer_circuit_data)
            .unwrap();

        let mut reader = Buffer::new(&serialized_receive_transfer_circuit_data);
        let decoded_receive_transfer_circuit = ReceiveTransferCircuit::from_buffer(
            &mut reader,
            &gate_serializer,
            &generator_serializer,
        )
        .unwrap();

        decoded_receive_transfer_circuit
    };

    if balance_processor
        .balance_transition_processor
        .receive_transfer_circuit
        .data
        != decoded_receive_transfer_circuit.data
    {
        panic!("mismatched receive_transfer_circuit data");
    }
    if balance_processor
        .balance_transition_processor
        .receive_transfer_circuit
        .target
        != decoded_receive_transfer_circuit.target
    {
        panic!("mismatched receive_transfer_circuit target");
    }

    println!("Complete receive_transfer_circuit");
}
