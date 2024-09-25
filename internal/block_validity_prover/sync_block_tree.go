package block_validity_prover

import (
	"encoding/hex"
	"errors"
	"fmt"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/block_builder_storage"
	bbsTypes "intmax2-node/internal/block_builder_storage/types"
	"intmax2-node/internal/block_post_service"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
	"time"

	"github.com/holiman/uint256"
)

func (p *blockValidityProver) BlockBuilder() block_builder_storage.BlockBuilderStorage {
	return p.blockBuilder
}

// func (p *blockValidityProver) SyncBlockTree(bps BlockSynchronizer) {
// 	blockContentChannel := make(chan A)
//
// 	go func() {
// 		for {
// 			startBlock, err := p.blockBuilder.LastSeenBlockPostedEventBlockNumber()
// 			if err != nil {
// 				panic(errors.Join(ErrLastSeenBlockPostedEventBlockNumberFail, err))
// 			}
//
// 			endBlock, err := p.syncBlockTree(bps, startBlock, blockContentChannel)
// 			if err != nil {
// 				panic(err)
// 			}
//
// 			err = p.blockBuilder.SetLastSeenBlockPostedEventBlockNumber(endBlock)
// 			if err != nil {
// 				var ErrSetLastSeenBlockPostedEventBlockNumberFail = errors.New("set last seen block posted event block number fail")
// 				panic(errors.Join(ErrSetLastSeenBlockPostedEventBlockNumberFail, err))
// 			}
//
// 			fmt.Printf("Block %d is searched\n", endBlock)
// 		}
// 	}()
//
// 	go func() {
// 		for {
// 			select {
// 			case <-p.ctx.Done():
// 				p.log.Warnf("Received cancel signal from context, stopping...")
// 				return
// 			case result := <-blockContentChannel:
// 				blockContent := result.blockContent
// 				postedBlock := result.postedBlock
//
// 				errChan <- p.syncBlockProver(blockContent, postedBlock)
// 			}
// 		}
// 	}()
// }

func (p *blockValidityProver) SyncBlockTree(db SQLDriverApp, bps BlockSynchronizer, startBlock uint64) (lastEventSeenBlockNumber uint64, err error) {
	var (
		events []*bindings.RollupBlockPosted
		nextBN *big.Int
	)

	latestScrollBlockNumber, err := bps.FetchLatestBlockNumber(p.ctx)
	if err != nil {
		return startBlock, errors.Join(ErrFetchLatestBlockNumberFail, err)
	}
	fmt.Printf("latestScrollBlockNumber: %d\n", latestScrollBlockNumber)

	const searchBlocksLimitAtOnce = 10000
	endBlock := startBlock + searchBlocksLimitAtOnce
	if endBlock > latestScrollBlockNumber {
		endBlock = latestScrollBlockNumber
	}

	events, nextBN, err = bps.FetchNewPostedBlocks(startBlock, &endBlock)
	if err != nil {
		return startBlock, errors.Join(ErrFetchNewPostedBlocksFail, err)
	}

	if len(events) == 0 {
		fmt.Printf("Scroll Block %d is synchronized (SyncBlockTree)\n", endBlock)
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

			// p.log.Debugf("blockContent: %v\n", blockContent)
			_, err = p.blockBuilder.CreateBlockContent(db, postedBlock, blockContent)
			if err != nil {
				panic(err)
				// return errors.Join(ErrCreateBlockContentFail, err)
			}
		}
	}

	return nextBN.Uint64(), nil
}

func (p *blockValidityProver) SyncBlockProverWithBlockNumber(
	db SQLDriverApp,
	blockNumber uint32,
) error {
	if blockNumber == 0 {
		return errors.New("genesis block number is not supported")
	}

	fmt.Printf("SyncBlockProverWithBlockNumber %d proof is synchronizing\n", blockNumber)
	result, err := p.blockBuilder.BlockAuxInfo(db, blockNumber)
	if err != nil {
		return err
	}

	return p.syncBlockProverWithAuxInfo(
		db,
		result.BlockContent,
		result.PostedBlock,
	)
}

func (p *blockValidityProver) syncBlockProverWithAuxInfo(
	db SQLDriverApp,
	blockContent *intMaxTypes.BlockContent,
	postedBlock *block_post_service.PostedBlock,
) error {
	fmt.Printf("IMPORTANT: Block %d proof is synchronizing\n", postedBlock.BlockNumber)

	blockWitness, err := p.blockBuilder.GenerateBlock(blockContent, postedBlock)
	if err != nil {
		panic(fmt.Errorf("failed to generate block: %w", err))
	}

	latestValidityWitness, err := p.blockBuilder.LastValidityWitness(db)
	if err != nil {
		panic(fmt.Errorf("failed to get last validity witness: %w", err))
	}

	fmt.Printf("blockWitness.Block.BlockNumber (syncBlockProverWithAuxInfo): %d\n", blockWitness.Block.BlockNumber)
	if blockWitness.Block.BlockNumber != p.blockBuilder.LatestIntMaxBlockNumber()+1 {
		fmt.Printf("db.LatestIntMaxBlockNumber(): %d\n", p.blockBuilder.LatestIntMaxBlockNumber())
		return errors.New("block number is not equal to the last block number + 1")
	}
	_, err = p.blockBuilder.CalculateValidityWitnessWithConsistencyCheck(blockWitness, latestValidityWitness)
	if err != nil {
		panic(fmt.Errorf("failed to calculate validity witness with consistency check: %w", err))
	}

	return p.syncBlockProver(db)
}

// lastPostedBlockNumber: block content generated
// lastBlockNumber: validity witness generated
// lastGeneratedBlockNumber: validity proof generated
// : balance proof generated

func (p *blockValidityProver) syncBlockProver(db SQLDriverApp) error {
	lastPostedBlockNumber, err := p.blockBuilder.LastPostedBlockNumber(db)
	if err != nil {
		panic(fmt.Errorf("failed to get last posted block number: %w", err))
	}

	var lastGeneratedBlockNumber uint32
	lastGeneratedBlockNumber, err = p.blockBuilder.LastGeneratedProofBlockNumber(db)
	if err != nil {
		return errors.Join(ErrLastGeneratedProofBlockNumberFail, err)
	}

	p.log.Debugf("lastGeneratedBlockNumber (SyncBlockProver): %d", lastGeneratedBlockNumber)
	p.log.Debugf("lastPostedBlockNumber (SyncBlockProver): %d", lastPostedBlockNumber)
	for blockNumber := lastGeneratedBlockNumber + 1; blockNumber <= lastPostedBlockNumber; blockNumber++ {
		p.log.Debugf("IMPORTANT: Block %d proof is processing", blockNumber)

		var validityWitness *bbsTypes.ValidityWitness
		validityWitness, err = p.blockBuilder.ValidityWitnessByBlockNumber(db, blockNumber)
		if err != nil {
			panic(fmt.Errorf("last validity witness error: %w", err))
		}

		p.log.Debugf("SenderFlag: %v", validityWitness.BlockWitness.Signature.SenderFlag)
		p.log.Debugf(
			"validityWitness.BlockWitness.Block.BlockNumber: %d",
			validityWitness.BlockWitness.Block.BlockNumber,
		)

		var lastValidityProof *string
		lastValidityProof, err = p.blockBuilder.ValidityProofByBlockNumber(db, blockNumber-1)
		if err != nil && !errors.Is(err, block_builder_storage.ErrGenesisValidityProof) {
			if err.Error() != ErrNoValidityProofByBlockNumber.Error() {
				return ErrNoValidityProofByBlockNumber
			}
		}

		p.log.Debugf(
			"validityWitness AccountRegistrationProofs: %v",
			validityWitness.ValidityTransitionWitness.AccountRegistrationProofs,
		)

		var validityProof string
		validityProof, err = p.requestAndFetchBlockValidityProof(validityWitness, lastValidityProof)
		if err != nil {
			return errors.Join(ErrRequestAndFetchBlockValidityProofFail, err)
		}

		var validityProofWithPis *intMaxTypes.Plonky2Proof
		validityProofWithPis, err = intMaxTypes.NewCompressedPlonky2ProofFromBase64String(validityProof)
		if err != nil {
			return errors.Join(ErrNewCompressedPlonky2ProofFromBase64StringFail, err)
		}

		validityPubicInputs := new(bbsTypes.ValidityPublicInputs).FromPublicInputs(validityProofWithPis.PublicInputs)
		p.log.Debugf(
			"SyncBlockProver block_proof block number: %d",
			validityPubicInputs.PublicState.BlockNumber,
		)
		p.log.Debugf(
			"SyncBlockProver block_proof prev account tree root: %s",
			validityPubicInputs.PublicState.PrevAccountTreeRoot.String(),
		)
		p.log.Debugf(
			"SyncBlockProver block_proof account tree root: %s",
			validityPubicInputs.PublicState.AccountTreeRoot.String(),
		)

		err = p.blockBuilder.SetValidityProof(db, validityWitness.BlockWitness.Block.Hash(), validityProof)
		if err != nil {
			panic(err)
		}

		p.log.Debugf("Block %d is reflected", validityWitness.BlockWitness.Block.BlockNumber)
	}

	return nil
}

func (p *blockValidityProver) requestAndFetchBlockValidityProof(
	validityWitness *bbsTypes.ValidityWitness,
	lastValidityProof *string,
) (validityProof string, err error) {
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
			validityProof, err = p.fetchBlockValidityProof(blockHash)
			if err != nil {
				continue
			}

			return validityProof, nil
		}
	}
}
