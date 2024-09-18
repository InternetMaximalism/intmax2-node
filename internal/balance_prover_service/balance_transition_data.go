package balance_prover_service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/balance_service"
	"intmax2-node/internal/block_validity_prover"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"log"
	"math/big"
	"sort"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/ffg"
)

type DepositDetails struct {
	Recipient         *intMaxAcc.PublicKey
	TokenIndex        uint32
	Amount            *big.Int
	Salt              *intMaxGP.PoseidonHashOut
	RecipientSaltHash common.Hash
	DepositID         uint32
	DepositHash       common.Hash
}

type BalanceTransitionData struct {
	Transactions []*intMaxTypes.TxDetails
	Deposits     []*DepositDetails
	Transfers    []*intMaxTypes.TransferDetailsWithProofBody
}

func NewBalanceTransitionData(ctx context.Context, cfg *configs.Config, userPrivateKey *intMaxAcc.PrivateKey) (*BalanceTransitionData, error) {
	intMaxWalletAddress := userPrivateKey.ToAddress()
	fmt.Printf("Starting balance prover service: %s\n", intMaxWalletAddress)

	storedTransitionData, err := balance_service.GetUserBalancesRawRequest(ctx, cfg, intMaxWalletAddress.String())
	if err != nil {
		const msg = "failed to get user all data: %+v"
		panic(fmt.Sprintf(msg, err.Error()))
	}

	decodedUserAllData, err := DecodeBackupData(ctx, cfg, storedTransitionData, userPrivateKey)
	if err != nil {
		fmt.Printf("Error in DecodeBackupData: %v\n", err)
		return nil, fmt.Errorf("failed to decode backup data: %v", err)
	}

	fmt.Printf("user deposits: %d\n", len(decodedUserAllData.Deposits))
	fmt.Printf("user transfers: %d\n", len(decodedUserAllData.Transfers))
	fmt.Printf("user transactions: %d\n", len(decodedUserAllData.Transactions))
	fmt.Println("Finished balance prover service")
	return decodedUserAllData, nil
}

func DecodeBackupData(
	ctx context.Context,
	cfg *configs.Config,
	userAllData *balance_service.GetBalancesResponse,
	userPrivateKey *intMaxAcc.PrivateKey,
) (*BalanceTransitionData, error) {
	receivedDeposits := make([]*DepositDetails, 0)
	receivedTransfers := make([]*intMaxTypes.TransferDetailsWithProofBody, 0)
	sentTransactions := make([]*intMaxTypes.TxDetails, 0)

	address := userPrivateKey.Public().ToAddress()
	fmt.Printf("Decoding backup data for address: %s\n", address.String())

	for _, deposit := range userAllData.Deposits {
		encryptedDepositBytes, err := base64.StdEncoding.DecodeString(deposit.EncryptedDeposit)
		if err != nil {
			log.Printf("failed to decode deposit: %v", err)
			continue
		}

		encodedDeposit, err := userPrivateKey.DecryptECIES(encryptedDepositBytes)
		if err != nil {
			log.Printf("failed to decrypt deposit: %v", err)
			continue
		}

		var decodedDeposit intMaxTypes.Deposit
		err = decodedDeposit.Unmarshal(encodedDeposit)
		if err != nil {
			log.Printf("failed to unmarshal deposit: %v", err)
			continue
		}

		// Request data store vault if deposit is valid
		depositIDStr := deposit.BlockNumber
		depositID, err := strconv.ParseUint(depositIDStr, 10, 32)
		for err != nil {
			log.Printf("failed to parse deposit ID: %v", err)
		}

		ok, err := balance_service.GetDepositValidityRawRequest(
			ctx,
			cfg,
			depositIDStr,
		)
		if err != nil {
			return nil, errors.Join(balance_service.ErrDepositValidity, err)
		}
		if !ok {
			continue
		}

		recipient, err := intMaxAcc.NewPublicKeyFromAddressHex(deposit.Recipient)
		if err != nil {
			log.Printf("failed to create recipient public key: %v", err)
			continue
		}

		recipientSaltHash := recipient.HashWithSalt(decodedDeposit.Salt)
		depositLeaf := intMaxTree.DepositLeaf{
			RecipientSaltHash: recipientSaltHash,
			TokenIndex:        decodedDeposit.TokenIndex,
			Amount:            decodedDeposit.Amount,
		}
		depositHash := depositLeaf.Hash()
		fmt.Printf("deposit ID: %v\n", depositID)
		fmt.Printf("deposit leaf: %v\n", depositLeaf)
		fmt.Printf("deposit (nullifier): %s\n", depositHash.String())
		deposit := DepositDetails{
			Recipient:         recipient,
			TokenIndex:        decodedDeposit.TokenIndex,
			Amount:            decodedDeposit.Amount,
			Salt:              decodedDeposit.Salt,
			RecipientSaltHash: recipientSaltHash,
			DepositID:         uint32(depositID),
			DepositHash:       depositHash,
		}

		receivedDeposits = append(receivedDeposits, &deposit)

		// if _, ok := balances[tokenIndex]; !ok {
		// 	balances[tokenIndex] = big.NewInt(0)
		// }

		// balances[tokenIndex] = new(big.Int).Add(balances[tokenIndex], decodedDeposit.Amount)
	}

	fmt.Printf("num of received transfers: %d\n", len(userAllData.Transfers))
	for i, transfer := range userAllData.Transfers {
		fmt.Printf("transfers[%d]: %v\n", i, transfer)
		encryptedTransferBytes, err := base64.StdEncoding.DecodeString(transfer.EncryptedTransfer)
		if err != nil {
			log.Printf("failed to decode transfer: %v", err)
			continue
		}
		fmt.Println("DecryptECIES")
		encodedTransfer, err := userPrivateKey.DecryptECIES(encryptedTransferBytes)
		if err != nil {
			log.Printf("failed to decrypt transfer: %v", err)
			continue
		}
		fmt.Println("Unmarshal")
		var decodedTransfer intMaxTypes.TransferDetails
		err = decodedTransfer.Unmarshal(encodedTransfer)
		if err != nil {
			log.Printf("failed to unmarshal transfer in DecodeBackupData: %v", err)
			continue
		}
		fmt.Println("end Unmarshal")

		transferWithProofBody := intMaxTypes.TransferDetailsWithProofBody{
			TransferDetails:                  &decodedTransfer,
			SenderLastBalanceProofBody:       transfer.SenderLastBalanceProofBody,
			SenderBalanceTransitionProofBody: transfer.SenderBalanceTransitionProofBody,
		}

		receivedTransfers = append(receivedTransfers, &transferWithProofBody)

		// if _, ok := balances[tokenIndex]; !ok {
		// 	balances[tokenIndex] = big.NewInt(0)
		// }

		// balances[tokenIndex] = new(big.Int).Add(balances[tokenIndex], decodedTransfer.Amount)
	}

	fmt.Printf("num of sent transactions: %d\n", len(userAllData.Transactions))
	for _, transaction := range userAllData.Transactions {
		fmt.Printf("transaction: %v\n", transaction)
		encryptedTxBytes, err := base64.StdEncoding.DecodeString(transaction.EncryptedTx)
		if err != nil {
			log.Printf("failed to decode transaction: %v", err)
			continue
		}
		encodedTx, err := userPrivateKey.DecryptECIES(encryptedTxBytes)
		if err != nil {
			log.Printf("failed to decrypt transaction: %v", err)
			continue
		}
		decodedTx, err := intMaxTypes.UnmarshalTxDetails(transaction.EncodingVersion, encodedTx)
		if err != nil {
			log.Printf("failed to unmarshal transaction: %v", err)
			continue
		}

		sentTransactions = append(sentTransactions, decodedTx)
		// for _, transfer := range decodedTx.Transfers {
		// 	if _, ok := balances[transfer.TokenIndex]; !ok {
		// 		balances[transfer.TokenIndex] = big.NewInt(0)
		// 	}
		// 	balances[transfer.TokenIndex] = new(big.Int).Sub(balances[transfer.TokenIndex], transfer.Amount)
		// }
	}

	return &BalanceTransitionData{
		Transactions: sentTransactions,
		Deposits:     receivedDeposits,
		Transfers:    receivedTransfers,
	}, nil
}

func (userAllData *BalanceTransitionData) SortValidUserData(
	log logger.Logger,
	blockValidityService block_validity_prover.BlockValidityService,
) ([]ValidBalanceTransition, error) {
	validDeposits, invalidDeposits, err := ExtractValidReceivedDeposits(log, userAllData, blockValidityService)
	if err != nil {
		fmt.Println("Error in ExtractValidReceivedDeposit")
	}
	fmt.Printf("num of valid deposits: %d\n", len(validDeposits))
	fmt.Printf("num of invalid deposits: %d\n", len(invalidDeposits))

	validTransfers, invalidTransfers, err := ExtractValidReceivedTransfers(log, userAllData, blockValidityService)
	if err != nil {
		fmt.Println("Error in ExtractValidReceivedDeposit")
	}
	fmt.Printf("num of valid transfers: %d\n", len(validTransfers))
	fmt.Printf("num of invalid transfers: %d\n", len(invalidTransfers))

	validTransactions, invalidTransactions, err := ExtractValidSentTransactions(log, userAllData, blockValidityService)
	if err != nil {
		fmt.Println("Error in ExtractValidReceivedDeposit")
	}
	fmt.Printf("num of valid transactions: %d\n", len(validTransactions))
	fmt.Printf("num of invalid transactions: %d\n", len(invalidTransactions))

	validBalanceTransitions := make([]ValidBalanceTransition, 0, len(validDeposits)+len(validTransfers)+len(validTransactions))
	for _, transaction := range validTransactions {
		validBalanceTransitions = append(validBalanceTransitions, transaction)
	}
	for _, deposit := range validDeposits {
		validBalanceTransitions = append(validBalanceTransitions, deposit)
	}
	for _, transfer := range validTransfers {
		validBalanceTransitions = append(validBalanceTransitions, transfer)
	}

	sort.Slice(validBalanceTransitions, func(i, j int) bool {
		return validBalanceTransitions[i].BlockNumber() < validBalanceTransitions[j].BlockNumber()
	})

	return validBalanceTransitions, nil
}

type ValidBalanceTransition interface {
	BlockNumber() uint32
}

type ValidSentTx struct {
	TxHash      *poseidonHashOut
	blockNumber uint32
	Tx          *intMaxTypes.TxDetails
}

func (v ValidSentTx) BlockNumber() uint32 {
	return v.blockNumber
}

type ValidReceivedDeposit struct {
	DepositHash common.Hash
	blockNumber uint32
	Deposit     *DepositDetails
}

func (v ValidReceivedDeposit) BlockNumber() uint32 {
	return v.blockNumber
}

type ValidReceivedTransfer struct {
	TransferHash *poseidonHashOut
	blockNumber  uint32
	Transfer     *intMaxTypes.TransferDetailsWithProofBody
}

func (v ValidReceivedTransfer) BlockNumber() uint32 {
	return v.blockNumber
}

func ExtractValidSentTransactions(
	log logger.Logger,
	userData *BalanceTransitionData,
	blockValidityService block_validity_prover.BlockValidityService,
) ([]ValidSentTx, []*poseidonHashOut, error) {
	sentBlockNumbers := make([]ValidSentTx, 0, len(userData.Deposits))
	invalidTxHashes := make([]*poseidonHashOut, 0, len(userData.Deposits))
	for _, tx := range userData.Transactions {
		txHash := tx.Tx.Hash() // TODO: validate transaction

		fmt.Printf("transaction hash: %s\n", txHash.String())
		if tx.TxTreeRoot == nil {
			// If TxTreeRoot is nil, the account is no longer valid.
			log.Warnf("transaction tx tree root is nil\n")
			invalidTxHashes = append(invalidTxHashes, txHash)
			continue
		}

		txRoot := tx.TxTreeRoot.String()[:2]
		blockContent, err := blockValidityService.BlockContentByTxRoot(txRoot)
		if err != nil {
			log.Warnf("failed to get block content by tx root %s: %v\n", txHash.String(), err)
			continue
		}

		blockNumber := blockContent.BlockNumber
		sentBlockNumbers = append(sentBlockNumbers, ValidSentTx{
			TxHash:      txHash,
			blockNumber: blockNumber,
			Tx:          tx,
		})

		log.Debugf("valid transaction: %s\n", txHash.String())
	}

	return sentBlockNumbers, invalidTxHashes, nil
}

func ExtractValidReceivedDeposits(
	log logger.Logger,
	userData *BalanceTransitionData,
	blockValidityService block_validity_prover.BlockValidityService,
) ([]ValidReceivedDeposit, []common.Hash, error) {
	sentBlockNumbers := make([]ValidReceivedDeposit, 0, len(userData.Deposits))
	invalidDepositHashes := make([]common.Hash, 0, len(userData.Deposits))
	for _, deposit := range userData.Deposits {
		defaultDepositHash := common.Hash{}
		if deposit.DepositHash == defaultDepositHash {
			log.Warnf("deposit hash should not be zero\n")
			continue
		}

		depositHash := deposit.DepositHash // TODO: validate deposit
		log.Debugf("deposit hash: %s\n", depositHash.String())

		depositInfo, err := blockValidityService.GetDepositInfoByHash(depositHash)
		if err != nil {
			log.Warnf("failed to get deposit index by hash %s: %v\n", depositHash.String(), err)
			continue
		}

		if depositInfo.BlockNumber == nil {
			log.Warnf("deposit block number is nil\n")
			continue
		}

		sentBlockNumbers = append(sentBlockNumbers, ValidReceivedDeposit{
			DepositHash: depositHash,
			blockNumber: *depositInfo.BlockNumber,
			Deposit:     deposit,
		})

		log.Debugf("valid deposit: %s\n", depositHash.String())
	}

	return sentBlockNumbers, invalidDepositHashes, nil
}

func ExtractValidReceivedTransfers(
	log logger.Logger,
	userData *BalanceTransitionData,
	blockValidityService block_validity_prover.BlockValidityService,
) ([]ValidReceivedTransfer, []*poseidonHashOut, error) {
	receivedBlockNumbers := make([]ValidReceivedTransfer, 0, len(userData.Transfers))
	invalidTransferHashes := make([]*poseidonHashOut, 0, len(userData.Transfers))
	for _, transfer := range userData.Transfers {
		transferHash := transfer.TransferDetails.TransferWitness.Transfer.Hash() // TODO: validate transfer

		log.Debugf("transfer hash: %s\n", transferHash.String())
		if transfer.TransferDetails.TxTreeRoot == nil {
			log.Warnf("transfer tx tree root is nil\n")
			invalidTransferHashes = append(invalidTransferHashes, transferHash)
			continue
		}

		txRoot := transfer.TransferDetails.TxTreeRoot.String()[:2]
		blockContent, err := blockValidityService.BlockContentByTxRoot(txRoot)
		if err != nil {
			log.Warnf("failed to get block content by transfer root %s: %v\n", transferHash.String(), err)
			continue
		}

		blockNumber := blockContent.BlockNumber
		receivedBlockNumbers = append(receivedBlockNumbers, ValidReceivedTransfer{
			TransferHash: transferHash,
			blockNumber:  blockNumber,
			Transfer:     transfer,
		})

		log.Debugf("valid transfer: %s\n", transferHash.String())
	}

	return receivedBlockNumbers, invalidTransferHashes, nil
}

// pub fn new(
// 	config: &CircuitConfig,
// 	circuit_type: BalanceTransitionType,
// 	receive_transfer_circuit: &ReceiveTransferCircuit<F, C, D>,
// 	receive_deposit_circuit: &ReceiveDepositCircuit<F, C, D>,
// 	update_circuit: &UpdateCircuit<F, C, D>,
// 	sender_circuit: &SenderCircuit<F, C, D>,
// 	receive_transfer_proof: Option<ProofWithPublicInputs<F, C, D>>,
// 	receive_deposit_proof: Option<ProofWithPublicInputs<F, C, D>>,
// 	update_proof: Option<ProofWithPublicInputs<F, C, D>>,
// 	sender_proof: Option<ProofWithPublicInputs<F, C, D>>,
// 	prev_balance_pis: BalancePublicInputs,
// 	balance_circuit_vd: VerifierOnlyCircuitData<C, D>,
// ) -> Self {
// 	let mut circuit_flags = [false; 4];
// 	circuit_flags[circuit_type as usize] = true;

// 	let new_balance_pis = match circuit_type {
// 		BalanceTransitionType::ReceiveTransfer => {
// 			let receive_transfer_proof = receive_transfer_proof
// 				.clone()
// 				.expect("receive_transfer_proof is None");
// 			receive_transfer_circuit
// 				.data
// 				.verify(receive_transfer_proof.clone())
// 				.expect("receive_transfer_proof is invalid");
// 			let pis = ReceiveTransferPublicInputs::<F, C, D>::from_slice(
// 				config,
// 				&receive_transfer_proof.public_inputs,
// 			);
// 			assert_eq!(
// 				pis.balance_circuit_vd, balance_circuit_vd,
// 				"balance_circuit_vd mismatch in receive_transfer_proof"
// 			);
// 			assert_eq!(
// 				pis.prev_private_commitment,
// 				prev_balance_pis.private_commitment,
// 			);
// 			assert_eq!(pis.pubkey, prev_balance_pis.pubkey);
// 			assert_eq!(pis.public_state, prev_balance_pis.public_state);
// 			BalancePublicInputs {
// 				pubkey: pis.pubkey,
// 				private_commitment: pis.new_private_commitment,
// 				..prev_balance_pis.clone()
// 			}
// 		}
// 		BalanceTransitionType::ReceiveDeposit => {
// 			let receive_deposit_proof = receive_deposit_proof
// 				.clone()
// 				.expect("receive_deposit_proof is None");
// 			receive_deposit_circuit
// 				.data
// 				.verify(receive_deposit_proof.clone())
// 				.expect("receive_deposit_proof is invalid");
// 			let pis = ReceiveDepositPublicInputs::from_u64_slice(
// 				&receive_deposit_proof
// 					.public_inputs
// 					.iter()
// 					.map(|x| x.to_canonical_u64())
// 					.collect::<Vec<_>>(),
// 			);
// 			assert_eq!(
// 				pis.prev_private_commitment,
// 				prev_balance_pis.private_commitment,
// 			);
// 			assert_eq!(pis.pubkey, prev_balance_pis.pubkey);
// 			assert_eq!(pis.public_state, prev_balance_pis.public_state);
// 			BalancePublicInputs {
// 				pubkey: pis.pubkey,
// 				private_commitment: pis.new_private_commitment,
// 				..prev_balance_pis.clone()
// 			}
// 		}
// 		BalanceTransitionType::Update => {
// 			let update_proof = update_proof.clone().expect("update_proof is None");
// 			update_circuit
// 				.data
// 				.verify(update_proof.clone())
// 				.expect("update_proof is invalid");
// 			let pis =
// 				UpdatePublicInputs::from_u64_slice(&update_proof.public_inputs.to_u64_vec());
// 			assert_eq!(pis.prev_public_state, prev_balance_pis.public_state);
// 			BalancePublicInputs {
// 				public_state: pis.new_public_state,
// 				..prev_balance_pis
// 			}
// 		}
// 		BalanceTransitionType::Sender => {
// 			let sender_proof = sender_proof.clone().expect("sender_proof is None");
// 			sender_circuit
// 				.data
// 				.verify(sender_proof.clone())
// 				.expect("sender_proof is invalid");
// 			let pis = SenderPublicInputs::from_u64_slice(
// 				&sender_proof
// 					.public_inputs
// 					.iter()
// 					.map(|x| x.to_canonical_u64())
// 					.collect::<Vec<_>>(),
// 			);
// 			assert_eq!(pis.prev_balance_pis, prev_balance_pis);
// 			pis.new_balance_pis
// 		}
// 	};

// 	let new_balance_pis_commitment = new_balance_pis.commitment();

type SenderPublicInputs struct {
	PrevBalancePublicInputs *BalancePublicInputs
	NewBalancePublicInputs  *BalancePublicInputs
}

func (pis *BalancePublicInputs) Equal(other *BalancePublicInputs) bool {
	return pis.PubKey.Equal(other.PubKey) &&
		pis.PrivateCommitment.Equal(other.PrivateCommitment) &&
		pis.PublicState.Equal(other.PublicState) &&
		pis.LastTxHash.Equal(other.LastTxHash) &&
		pis.LastTxInsufficientFlags.Equal(&other.LastTxInsufficientFlags)
}

func (pis *SenderPublicInputs) FromPublicInputs(
	publicInputs []ffg.Element,
) (*SenderPublicInputs, error) {
	prevBalancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(publicInputs[:sizeOfBalancePublicInputs])
	if err != nil {
		return nil, err
	}

	newBalancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(publicInputs[sizeOfBalancePublicInputs:])
	if err != nil {
		return nil, err
	}

	return &SenderPublicInputs{
		PrevBalancePublicInputs: prevBalancePublicInputs,
		NewBalancePublicInputs:  newBalancePublicInputs,
	}, nil
}

func (pis *BalancePublicInputs) UpdateWithSendTransition(
	balanceTransitionPublicInputs *SenderPublicInputs,
) (*BalancePublicInputs, error) {
	if !balanceTransitionPublicInputs.PrevBalancePublicInputs.Equal(pis) {
		return nil, errors.New("prev balance public inputs mismatch")
	}

	return balanceTransitionPublicInputs.NewBalancePublicInputs, nil
}
