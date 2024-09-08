package block_validity_prover

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/block_post_service"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
	"time"

	"github.com/holiman/uint256"
)

func (ai *mockBlockBuilder) RegisterPublicKey(pk *intMaxAcc.PublicKey, lastSentBlockNumber uint32) (accountID uint64, err error) {
	publicKey := pk.BigInt()
	proof, _, err := ai.AccountTree.ProveMembership(publicKey)
	if err != nil {
		var ErrProveMembershipFail = errors.New("failed to prove membership")
		return 0, errors.Join(ErrProveMembershipFail, err)
	}

	_, ok := ai.AccountTree.GetAccountID(publicKey)
	if ok {
		_, err = ai.AccountTree.Update(publicKey, uint64(0))
		if err != nil {
			var ErrUpdateAccountFail = errors.New("failed to update account")
			return 0, errors.Join(ErrUpdateAccountFail, err)
		}

		return uint64(proof.LeafIndex), nil
	}

	var insertionProof *intMaxTree.IndexedInsertionProof
	insertionProof, err = ai.AccountTree.Insert(publicKey, uint64(0))
	if err != nil {
		return 0, errors.Join(ErrCreateAccountFail, err)
	}

	return uint64(insertionProof.Index), nil
}

func (ai *mockBlockBuilder) PublicKeyByAccountID(accountID uint64) (pk *intMaxAcc.PublicKey, err error) {
	var accID uint256.Int
	accID.SetUint64(accountID)

	acc := ai.AccountTree.GetLeaf(accountID)

	pk, err = new(intMaxAcc.PublicKey).SetBigInt(acc.Key)
	if err != nil {
		return nil, errors.Join(ErrDecodeHexToPublicKeyFail, err)
	}

	return pk, nil
}

func (ai *mockBlockBuilder) AccountBySenderAddress(_ string) (*uint256.Int, error) {
	return nil, fmt.Errorf("AccountBySenderAddress not implemented")
}

func (p *blockValidityProver) BlockBuilder() *mockBlockBuilder {
	return p.blockBuilder
}

// func (p *blockValidityProver) SyncBlockTree(bps BlockSynchronizer) {
// 	blockContentChannel := make(chan A)

// 	go func() {
// 		for {
// 			startBlock, err := p.blockBuilder.LastSeenBlockPostedEventBlockNumber()
// 			if err != nil {
// 				panic(errors.Join(ErrLastSeenBlockPostedEventBlockNumberFail, err))
// 			}

// 			endBlock, err := p.syncBlockTree(bps, startBlock, blockContentChannel)
// 			if err != nil {
// 				panic(err)
// 			}

// 			err = p.blockBuilder.SetLastSeenBlockPostedEventBlockNumber(endBlock)
// 			if err != nil {
// 				var ErrSetLastSeenBlockPostedEventBlockNumberFail = errors.New("set last seen block posted event block number fail")
// 				panic(errors.Join(ErrSetLastSeenBlockPostedEventBlockNumberFail, err))
// 			}

// 			fmt.Printf("Block %d is searched\n", endBlock)
// 		}
// 	}()

// 	go func() {
// 		for {
// 			select {
// 			case <-p.ctx.Done():
// 				p.log.Warnf("Received cancel signal from context, stopping...")
// 				return
// 			case result := <-blockContentChannel:
// 				blockContent := result.blockContent
// 				postedBlock := result.postedBlock

// 				errChan <- p.syncBlockProver(blockContent, postedBlock)
// 			}
// 		}
// 	}()
// }

func (p *blockValidityProver) SyncBlockTree(bps BlockSynchronizer, startBlock uint64) (lastEventSeenBlockNumber uint64, err error) {
	var (
		events []*bindings.RollupBlockPosted
		nextBN *big.Int
	)

	latestScrollBlockNumber, err := bps.FetchLatestBlockNumber(p.ctx)

	const int5000Key = 5000
	endBlock := startBlock + int5000Key
	if endBlock > latestScrollBlockNumber {
		endBlock = latestScrollBlockNumber
	}

	events, nextBN, err = bps.FetchNewPostedBlocks(startBlock, &endBlock)
	if err != nil {
		return startBlock, errors.Join(ErrFetchNewPostedBlocksFail, err)
	}

	if len(events) == 0 {
		return endBlock, nil
	}

	tickerEventWatcher := time.NewTicker(p.cfg.BlockValidityProver.TimeoutForEventWatcher)
	defer func() {
		if tickerEventWatcher != nil {
			tickerEventWatcher.Stop()
		}
	}()

	for key := range events {
		select {
		case <-p.ctx.Done():
			p.log.Warnf("Received cancel signal from context, stopping...")
			return startBlock, p.ctx.Err()
		case <-tickerEventWatcher.C:
			fmt.Println("tickerEventWatcher.C")
			var blN uint256.Int
			_ = blN.SetFromBig(new(big.Int).SetUint64(events[key].Raw.BlockNumber))

			time.Sleep(1 * time.Second)

			var cd []byte
			cd, err = bps.FetchScrollCalldataByHash(events[key].Raw.TxHash)
			if err != nil {
				return startBlock, errors.Join(ErrFetchScrollCalldataByHashFail, err)
			}

			postedBlock := block_post_service.NewPostedBlock(
				events[key].PrevBlockHash,
				events[key].DepositTreeRoot,
				uint32(events[key].BlockNumber.Uint64()),
				events[key].SignatureHash,
			)

			intMaxBlockNumber := events[key].BlockNumber

			// Update account tree
			var blockContent *intMaxTypes.BlockContent
			blockContent, err = FetchIntMaxBlockContentByCalldata(cd, postedBlock, p.blockBuilder)
			if err != nil {
				err = errors.Join(ErrFetchIntMaxBlockContentByCalldataFail, err)
				switch {
				case errors.Is(err, ErrUnknownAccountID):
					const msg = "block %q is ErrUnknownAccountID"
					p.log.WithError(err).Errorf(msg, intMaxBlockNumber.String())
				case errors.Is(err, ErrCannotDecodeAddress):
					const msg = "block %q is ErrCannotDecodeAddress"
					p.log.WithError(err).Errorf(msg, intMaxBlockNumber.String())
				default:
					const msg = "block %q processing error occurred"
					p.log.WithError(err).Errorf(msg, intMaxBlockNumber.String())
				}

				const msg = "processing of block %q error occurred"
				p.log.Debugf(msg, intMaxBlockNumber.String())
			} else {
				const msg = "block %q is found (Scroll block number: %s)"
				p.log.Debugf(msg, intMaxBlockNumber.String(), blN.String())
			}

			senders := make([]intMaxTypes.ColumnSender, len(blockContent.Senders))
			for i, sender := range blockContent.Senders {
				senders[i] = intMaxTypes.ColumnSender{
					PublicKey: hex.EncodeToString(sender.PublicKey.ToAddress().Bytes()),
					AccountID: sender.AccountID,
					IsSigned:  sender.IsSigned,
				}
			}

			err := setAuxInfo(
				p.blockBuilder,
				postedBlock,
				blockContent,
			)
			if err != nil {
				panic(err)
				// return errors.Join(ErrCreateBlockContentFail, err)
			}
		}
	}

	return nextBN.Uint64(), nil
}

func (p *blockValidityProver) SyncBlockProverWithAuxInfo(
	blockContent *intMaxTypes.BlockContent,
	postedBlock *block_post_service.PostedBlock,
) error {
	fmt.Printf("IMPORTANT: Block %d proof is synchronizing\n", postedBlock.BlockNumber)

	// TODO: Update block hash tree

	// TODO: Separate another worker

	blockWitness, err := p.blockBuilder.GenerateBlock(blockContent, postedBlock)
	if err != nil {
		panic(err)
	}

	latestValidityWitness, err := p.blockBuilder.LastValidityWitness()
	if err != nil {
		panic("last validity witness error")
	}

	validityWitness, err := generateValidityWitness(p.blockBuilder, blockWitness, latestValidityWitness)
	if err != nil {
		panic(err)
	}

	if err := p.blockBuilder.SetValidityWitness(
		validityWitness.BlockWitness.Block.BlockNumber,
		validityWitness,
	); err != nil {
		panic(err)
	}

	return p.SyncBlockProver()
}

// lastPostedBlockNumber: block content generated
// lastBlockNumber: validity witness generated
// lastGeneratedBlockNumber: validity proof generated
// : balance proof generated

func (p *blockValidityProver) SyncBlockProver() error {
	currentBlockNumber := p.blockBuilder.LatestIntMaxBlockNumber()
	lastGeneratedBlockNumber := p.blockBuilder.LastGeneratedProofBlockNumber
	// if lastGeneratedBlockNumber >= currentBlockNumber {
	// 	fmt.Printf("Block %d is already done\n", lastGeneratedBlockNumber)
	// 	fmt.Printf("prepared witness\n", currentBlockNumber)
	// 	lastPostedBlockNumber := p.BlockBuilder().LastPostedBlockNumber
	// 	fmt.Printf("last posted number is %d\n", lastPostedBlockNumber)
	// 	return nil
	// }

	for blockNumber := lastGeneratedBlockNumber + 1; blockNumber <= currentBlockNumber; blockNumber++ {
		// validityWitnessBlockNumber := p.blockBuilder.LatestIntMaxBlockNumber()
		validityWitness, err := p.blockBuilder.ValidityWitnessByBlockNumber(blockNumber)
		fmt.Printf("SenderFlag: %v\n", validityWitness.BlockWitness.Signature.SenderFlag)

		// validityWitnessBlockNumber := validityWitness.BlockWitness.Block.BlockNumber
		fmt.Printf("IMPORTANT: Block %d proof is processing\n", validityWitness.BlockWitness.Block.BlockNumber)
		fmt.Printf("IMPORTANT: Block %d proof is processing\n", blockNumber)
		if err != nil {
			panic("last validity witness error")
		}

		encodedBlockWitness, err := json.Marshal(validityWitness.BlockWitness)
		if err != nil {
			panic("marshal validity witness error")
		}
		fmt.Printf("encodedBlockWitness (SyncBlockProver): %s\n", encodedBlockWitness)

		lastValidityProof, err := p.blockBuilder.ValidityProofByBlockNumber(blockNumber - 1)
		if err != nil && err.Error() != ErrGenesisValidityProof.Error() {
			var ErrLastValidityProofFail = errors.New("last validity proof fail")
			return errors.Join(ErrLastValidityProofFail, err)
		}

		fmt.Printf("validityWitness AccountRegistrationProofs: %v\n", validityWitness.ValidityTransitionWitness.AccountRegistrationProofs)

		validityProof, err := p.requestAndFetchBlockValidityProof(validityWitness, lastValidityProof)
		if err != nil {
			return errors.Join(ErrRequestAndFetchBlockValidityProofFail, err)
		}

		validityProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(validityProof)
		if err != nil {
			var ErrNewCompressedPlonky2ProofFromBase64StringFail = errors.New("new compressed plonky2 proof from base64 string fail")
			return errors.Join(ErrNewCompressedPlonky2ProofFromBase64StringFail, err)
		}
		validityPubicInputs := new(ValidityPublicInputs).FromPublicInputs(validityProofWithPis.PublicInputs)
		fmt.Printf("SyncBlockProver block_proof block number: %d\n", validityPubicInputs.PublicState.BlockNumber)
		fmt.Printf("SyncBlockProver block_proof prev account tree root: %s\n", validityPubicInputs.PublicState.PrevAccountTreeRoot.String())
		fmt.Printf("SyncBlockProver block_proof account tree root: %s\n", validityPubicInputs.PublicState.AccountTreeRoot.String())

		err = p.blockBuilder.SetValidityProof(validityWitness.BlockWitness.Block.BlockNumber, validityProof)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Block %d is reflected\n", validityWitness.BlockWitness.Block.BlockNumber)
	}

	return nil
}

func (p *blockValidityProver) requestAndFetchBlockValidityProof(validityWitness *ValidityWitness, lastValidityProof *string) (validityProof string, err error) {
	blockHash := validityWitness.BlockWitness.Block.Hash()
	blockNumber := validityWitness.BlockWitness.Block.BlockNumber
	fmt.Printf("Block %d is requested\n", blockNumber)

	err = p.requestBlockValidityProof(blockHash, validityWitness, lastValidityProof)
	if err != nil {
		var ErrRequestBlockValidityProofFail = errors.New("request block validity proof fail")
		return "", errors.Join(ErrRequestBlockValidityProofFail, err)
	}
	// last validity proof fail
	tickerBlockValidityProof := time.NewTicker(p.cfg.BlockValidityProver.TimeoutForFetchingBlockValidityProof)
	defer func() {
		if tickerBlockValidityProof != nil {
			tickerBlockValidityProof.Stop()
		}
	}()
	for {
		select {
		case <-p.ctx.Done():
			return "", p.ctx.Err()
		case <-tickerBlockValidityProof.C:
			var validityProof string
			validityProof, err = p.fetchBlockValidityProof(blockHash)
			if err != nil {
				continue
			}

			return validityProof, nil
		}
	}
}