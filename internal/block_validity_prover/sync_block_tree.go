package block_validity_prover

import (
	"context"
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/intmax_block_content"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	"math/big"
	"strings"
	"time"

	"github.com/holiman/uint256"
)

func (ai *mockBlockBuilder) RegisterPublicKey(_ *intMaxAcc.PublicKey) error {
	return fmt.Errorf("RegisterPublicKey not implemented")
}

func (ai *mockBlockBuilder) PublicKeyByAccountID(blockNumber uint32, accountID uint64) (pk *intMaxAcc.PublicKey, err error) {
	// var accID uint256.Int
	// accID.SetUint64(accountID)

	merkleTreeHistory, ok := ai.MerkleTreeHistory.MerkleTrees[blockNumber]
	if !ok {
		return nil, errors.New("merkle tree not found")
	}

	accountTree := merkleTreeHistory.AccountTree
	acc := accountTree.GetLeaf(accountID)

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

func (p *blockValidityProver) SyncBlockContent() (lastEventSeenBlockNumber uint64, err error) {
	err = p.blockBuilder.Exec(p.ctx, &lastEventSeenBlockNumber, func(d interface{}, in interface{}) error {
		q, _ := d.(SQLDriverApp)

		nextBN, ok := in.(*uint64)
		if !ok {
			const msg = "error convert value to *uint32"
			return fmt.Errorf(msg)
		}

		var startBlock uint64

		var event *mDBApp.EventBlockNumberForValidityProver
		event, err = q.EventBlockNumberByEventNameForValidityProver(mDBApp.BlockPostedEvent)
		switch {
		case err != nil:
			if !errors.Is(err, errorsDB.ErrNotFound) {
				const msg = "error fetching event block number: %w"
				return fmt.Errorf(msg, err)
			}

			fallthrough
		default:
			var needReset bool
			if event == nil {
				needReset = true
			} else {
				startBlock = event.LastProcessedBlockNumber
			}
			if needReset || startBlock == 0 {
				err = q.DelAllBlockAccounts()
				if err != nil {
					return errors.Join(ErrDelAllAccountsFail, err)
				}

				err = q.ResetSequenceByBlockAccounts()
				if err != nil {
					return errors.Join(ErrResetSequenceByAccountsFail, err)
				}

				startBlock = p.cfg.Blockchain.RollupContractDeployedBlockNumber
			}
		}

		p.log.Debugf("startBlock of LastSeenBlockPostedEventBlockNumber: %d", startBlock)

		*nextBN = startBlock

		defer func() {
			if err == nil {
				p.log.Debugf("endBlock of LastSeenBlockPostedEventBlockNumber: %d", *nextBN)
			}
		}()

		var latestScrollBlockNumber uint64
		latestScrollBlockNumber, err = p.blockSynchronizer.FetchLatestBlockNumber(p.ctx)
		if err != nil {
			return errors.Join(ErrFetchLatestBlockNumberFail, err)
		}
		p.log.Debugf("latestScrollBlockNumber: %d", latestScrollBlockNumber)

		const searchBlocksLimitAtOnce = 100
		endBlock := startBlock + searchBlocksLimitAtOnce
		if endBlock > latestScrollBlockNumber {
			endBlock = latestScrollBlockNumber
		}

		var events []*bindings.RollupBlockPosted
		events, _, err = p.blockSynchronizer.FetchNewPostedBlocks(startBlock, &endBlock)
		if err != nil {
			return errors.Join(ErrFetchNewPostedBlocksFail, err)
		}

		if len(events) == 0 {
			*nextBN = endBlock
			p.log.Debugf("Scroll Block %d is synchronized (SyncBlockTree)", endBlock)

			_, err = q.UpsertEventBlockNumberForValidityProver(mDBApp.BlockPostedEvent, endBlock)
			if err != nil {
				return err
			}

			return nil
		}

		tickerEventWatcher := time.NewTicker(200 * time.Millisecond)
		defer func() {
			if tickerEventWatcher != nil {
				tickerEventWatcher.Stop()
			}
		}()

		var curBlN uint256.Int
		_ = curBlN.SetUint64(startBlock)

		var useTickerEventWatcher bool
		for key := range events {
			select {
			case <-p.ctx.Done():
				p.log.Warnf("Received cancel signal from context, stopping...")
				*nextBN = curBlN.Uint64()
				return nil
			case <-tickerEventWatcher.C:
				if useTickerEventWatcher {
					continue
				}

				useTickerEventWatcher = true

				_, err = syncBlockContentWithEvent(p.ctx, p.log, q, p.blockBuilder, p.blockSynchronizer, events[key])
				if err != nil {
					if strings.HasPrefix(err.Error(), "copy account tree error") {
						return nil
					}

					return errors.Join(ErrProcessingBlocksFail, err)
				}

				_, err = q.UpsertEventBlockNumberForValidityProver(mDBApp.BlockPostedEvent, events[key].Raw.BlockNumber)
				if err != nil {
					return err
				}

				*nextBN = events[key].Raw.BlockNumber

				useTickerEventWatcher = false
			}
		}

		return nil
	})

	return lastEventSeenBlockNumber, err
}

func syncBlockContentWithEvent(
	ctx context.Context,
	log logger.Logger,
	db SQLDriverApp,
	blockBuilder *mockBlockBuilder,
	bps BlockSynchronizer,
	event *bindings.RollupBlockPosted,
) (*mDBApp.BlockContentWithProof, error) {
	intMaxBlockNumber := uint32(event.BlockNumber.Uint64())
	newBlockContent, err := db.BlockContentByBlockNumber(intMaxBlockNumber)
	if err == nil {
		return newBlockContent, nil
	} else if !errors.Is(err, errorsDB.ErrNotFound) {
		return nil, errors.Join(ErrProcessingBlocksFail, err)
	}

	var blN uint256.Int
	_ = blN.SetFromBig(new(big.Int).SetUint64(event.Raw.BlockNumber))

	var cd []byte
	cd, err = bps.FetchScrollCalldataByHash(event.Raw.TxHash)
	if err != nil {
		return nil, errors.Join(ErrFetchScrollCalldataByHashFail, err)
	}

	postedBlock := intmax_block_content.NewPostedBlock(
		event.PrevBlockHash,
		event.DepositTreeRoot,
		intMaxBlockNumber,
		event.SignatureHash,
	)

	// Update account tree
	var blockContent *intMaxTypes.BlockContent
	blockContent, err = intmax_block_content.FetchIntMaxBlockContentByCalldata(cd, postedBlock, blockBuilder)
	if err == nil {
		err = blockContent.IsValid()
	}
	if err != nil {
		err = errors.Join(ErrFetchIntMaxBlockContentByCalldataFail, err)
		switch {
		case errors.Is(err, intmax_block_content.ErrUnknownAccountID):
			const msg = "block %d is ErrUnknownAccountID"
			log.WithError(err).Errorf(msg, intMaxBlockNumber)
		case errors.Is(err, intmax_block_content.ErrCannotDecodeAddress):
			const msg = "block %d is ErrCannotDecodeAddress"
			log.WithError(err).Errorf(msg, intMaxBlockNumber)
		case errors.Is(err, intMaxAcc.ErrInvalidBlockSignature):
			const msg = "block %d is ErrInvalidBlockSignature"
			log.WithError(err).Errorf(msg, intMaxBlockNumber)
		default:
			const msg = "block %d processing error occurred"
			log.WithError(err).Errorf(msg, intMaxBlockNumber)
		}

		const msg = "processing of block %d error occurred"
		log.Debugf(msg, intMaxBlockNumber)
	} else {
		const msg = "block %d is found (SyncBlockTree, Scroll block number: %d)"
		log.Debugf(msg, intMaxBlockNumber, blN)
	}

	newBlockContent, err = blockBuilder.CreateBlockContent(
		ctx, db, postedBlock, blockContent, &blN, event.Raw.BlockHash)
	if err != nil {
		return nil, errors.Join(ErrCreateBlockContentFail, err)
	}

	blockWitness, err := blockBuilder.GenerateBlockWithTxTreeFromBlockContent(
		blockContent,
		postedBlock,
	)
	if err != nil {
		return nil, err
	}

	validityWitness, _, _, err := calculateValidityWitness(blockBuilder, blockWitness)
	if err != nil {
		return nil, err
	}

	_, invalidReason := validityWitness.BlockWitness.MainValidationPublicInputs()
	if invalidReason != "" {
		log.Debugf("invalid reason: %v\n", invalidReason)
	}

	return newBlockContent, nil
}

// func (p *blockValidityProver) SyncBlockValidityWitness() error {
// 	lastValidityWitnessBlockNumber := uint32(0)
// 	lastValidityWitness, err := p.blockBuilder.LastValidityWitness()
// 	if err != nil {
// 		if err.Error() != "not found" {
// 			var ErrLastValidityWitnessFail = errors.New("last validity witness fail")
// 			return errors.Join(ErrLastValidityWitnessFail, err)
// 		}
// 	} else {
// 		lastValidityWitnessBlockNumber = lastValidityWitness.BlockWitness.Block.BlockNumber
// 	}

// 	timeout := 5 * time.Second
// 	tickerValidityProver := time.NewTicker(timeout)
// 	defer func() {
// 		if tickerValidityProver != nil {
// 			tickerValidityProver.Stop()
// 		}
// 	}()

// 	for {
// 		select {
// 		case <-p.ctx.Done():
// 			p.log.Warnf("Received cancel signal from context, stopping...")
// 			return p.ctx.Err()
// 		case <-tickerValidityProver.C:
// 			fmt.Println("tickerValidityProver.C")
// 			err = p.generateBlockValidityWitness(lastValidityWitnessBlockNumber + 1)
// 			if err != nil {
// 				if errors.Is(err, ErrBlockContentByBlockNumber) || errors.Is(err, ErrRootBlockNumberNotFound) {
// 					continue
// 				}

// 				return err
// 			}

// 			lastValidityWitnessBlockNumber++
// 		}
// 	}
// }

// func (p *blockValidityProver) generateBlockValidityWitness(validityWitnessBlockNumber uint32) error {
// 	blockContent, err := p.blockBuilder.db.BlockContentByBlockNumber(validityWitnessBlockNumber)
// 	if err != nil {
// 		if err.Error() != "not found" {
// 			return errors.Join(ErrProcessingBlocksFail, err)
// 		}

// 		p.log.Warnf("WARNING: block content %d is not found\n", validityWitnessBlockNumber)
// 		return ErrBlockContentByBlockNumber
// 		// continue
// 	}
// 	fmt.Printf("====== generateValidityWitness: block %d ========\n", blockContent.BlockNumber)

// 	auxInfo, err := blockAuxInfoFromBlockContent(blockContent)
// 	if err != nil {
// 		return errors.Join(ErrProcessingBlocksFail, err)
// 	}

// 	// prevBlockNumber := uint32(events[key].BlockNumber.Uint64()) - 1
// 	prevBlockNumber := blockContent.BlockNumber - 1
// 	prevValidityWitness, err := p.blockBuilder.ValidityWitnessByBlockNumber(prevBlockNumber)
// 	if err != nil {
// 		return fmt.Errorf("failed to get last validity witness: %w", err)
// 		// panic(err)
// 	}

// 	_, err = p.UpdateValidityWitness(auxInfo.BlockContent, prevValidityWitness)
// 	if err != nil {
// 		if errors.Is(err, ErrRootBlockNumberNotFound) {
// 			fmt.Printf("WARNING: root block number %d not found\n", blockContent.BlockNumber)
// 			return ErrRootBlockNumberNotFound
// 			// continue
// 		}

// 		return fmt.Errorf("failed to update validity witness: %w", err)
// 		// panic(err)
// 	}

// 	return nil
// }

func (p *blockValidityProver) SyncBlockValidityProof() error {
	lastGeneratedProofBlockNumber, err := p.blockBuilder.LastGeneratedProofBlockNumber()
	if err != nil {
		var ErrLastGeneratedProofBlockNumberFail = errors.New("last generated proof block number fail")
		return errors.Join(ErrLastGeneratedProofBlockNumberFail, err)
	}

	timeout := 5 * time.Second
	tickerValidityProver := time.NewTicker(timeout)
	defer func() {
		if tickerValidityProver != nil {
			tickerValidityProver.Stop()
		}
	}()

	for {
		select {
		case <-p.ctx.Done():
			p.log.Warnf("Received cancel signal from context, stopping...")
			return p.ctx.Err()
		case <-tickerValidityProver.C:
			fmt.Println("tickerValidityProver.C")
			if err := p.generateValidityProof(lastGeneratedProofBlockNumber + 1); err != nil {
				fmt.Printf("generateValidityProof error (syncBlockValidityProof): %v\n", err)
				if errors.Is(err, ErrBlockContentByBlockNumber) || errors.Is(err, ErrRootBlockNumberNotFound) {
					continue
				}

				return err
			}

			fmt.Printf("Block %d is done (syncBlockValidityProof)\n", lastGeneratedProofBlockNumber+1)
			lastGeneratedProofBlockNumber++
		}
	}
}

// func (p *blockValidityProver) SyncBlockProverWithBlockNumber(
// 	blockNumber uint32,
// ) error {
// 	if blockNumber == 0 {
// 		return errors.New("genesis block number is not supported")
// 	}

// 	fmt.Printf("SyncBlockProverWithBlockNumber %d proof is synchronizing\n", blockNumber)
// 	result, err := p.blockBuilder.BlockAuxInfo(blockNumber)
// 	if err != nil {
// 		return err
// 	}

// 	return p.syncBlockProverWithAuxInfo(
// 		result.BlockContent,
// 		result.PostedBlock,
// 	)
// }

// func (p *blockValidityProver) syncBlockProverWithAuxInfo(
// 	blockContent *intMaxTypes.BlockContent,
// 	postedBlock *block_post_service.PostedBlock,
// ) error {
// 	fmt.Printf("IMPORTANT: Block %d proof is synchronizing\n", postedBlock.BlockNumber)

// 	blockWitness, err := p.blockBuilder.GenerateBlock(blockContent, postedBlock)
// 	if err != nil {
// 		panic(fmt.Errorf("failed to generate block: %w", err))
// 	}

// 	fmt.Printf("blockWitness.AccountMembershipProofs (syncBlockProverWithAuxInfo): %v\n", blockWitness.AccountMembershipProofs.IsSome)
// 	latestValidityWitness, _, err := calculateValidityWitness(p.blockBuilder, blockWitness)
// 	if err != nil {
// 		if errors.Is(err, ErrRootBlockNumberNotFound) {
// 			return ErrRootBlockNumberNotFound
// 		}

// 		panic(fmt.Errorf("failed to get last validity witness: %w", err))
// 	}

// 	fmt.Printf("blockWitness.Block.BlockNumber (syncBlockProverWithAuxInfo): %d\n", blockWitness.Block.BlockNumber)

// 	{
// 		// fmt.Printf("block.BlockNumber: %d\n", blockWitness.Block.BlockNumber)
// 		fmt.Printf("prevBlockTreeRoot: %s\n", blockWitness.PrevBlockTreeRoot.String())
// 		for i, blockHashes := range p.blockBuilder.BlockTree.Leaves {
// 			fmt.Printf("block tree leaves[%d]: %x\n", i, blockHashes.Marshal())
// 		}
// 		defaultBlockLeafHash := new(intMaxTree.BlockHashLeaf).SetDefault()
// 		fmt.Printf("block tree default leaf: %x\n", defaultBlockLeafHash.Marshal())
// 		fmt.Printf("block tree default leaf hash: %x\n", defaultBlockLeafHash.Hash().Marshal())
// 		for i, sibling := range latestValidityWitness.ValidityTransitionWitness.BlockMerkleProof.Siblings {
// 			fmt.Printf("validity transition sibling[%d]: %s\n", i, sibling.String())
// 		}
// 	}
// 	fmt.Printf("latestValidityWitness.BlockWitness.Block.BlockNumber: %d\n", latestValidityWitness.BlockWitness.Block.BlockNumber)
// 	// prevPis := latestValidityWitness.ValidityPublicInputs()
// 	// if blockWitness.Block.BlockNumber != prevPis.PublicState.BlockNumber+1 {
// 	// 	fmt.Printf("latestValidityWitness.BlockWitness.Block.BlockNumber: %d\n", latestValidityWitness.BlockWitness.Block.BlockNumber)
// 	// 	return errors.New("block number is not equal to the last block number + 1")
// 	// }
// 	// _, err = calculateValidityWitness(p.blockBuilder, blockWitness, latestValidityWitness)
// 	// if err != nil {
// 	// 	panic(fmt.Errorf("failed to calculate validity witness with consistency check: %w", err))
// 	// }

// 	return p.syncBlockProver()
// }

// lastPostedBlockNumber: block content generated
// lastBlockNumber: validity witness generated
// lastGeneratedBlockNumber: validity proof generated
// : balance proof generated

// func (p *blockValidityProver) syncBlockProver() error {
// 	lastPostedBlockNumber, err := p.blockBuilder.db.LastPostedBlockNumber()
// 	if err != nil {
// 		panic(fmt.Errorf("failed to get last posted block number: %w", err))
// 	}

// 	// validityProverInfo, err := p.FetchValidityProverInfo()
// 	// if err != nil {
// 	// 	var ErrFetchValidityProverInfoFail = errors.New("fetch validity prover info fail")
// 	// 	return errors.Join(ErrFetchValidityProverInfoFail, err)
// 	// }

// 	// currentBlockNumber := validityProverInfo.BlockNumber
// 	lastGeneratedBlockNumber, err := p.blockBuilder.LastGeneratedProofBlockNumber()
// 	if err != nil {
// 		if err.Error() != "not found" {
// 			var ErrLastGeneratedProofBlockNumberFail = errors.New("last generated proof block number fail")
// 			return errors.Join(ErrLastGeneratedProofBlockNumberFail, err)
// 		}

// 		lastGeneratedBlockNumber = 0
// 	}
// 	// if lastGeneratedBlockNumber >= currentBlockNumber {
// 	// 	fmt.Printf("Block %d is already done\n", lastGeneratedBlockNumber)
// 	// 	fmt.Printf("prepared witness\n", currentBlockNumber)
// 	// 	lastPostedBlockNumber := p.BlockBuilder().LastPostedBlockNumber
// 	// 	fmt.Printf("last posted number is %d\n", lastPostedBlockNumber)
// 	// 	return nil
// 	// }

// 	fmt.Printf("lastGeneratedBlockNumber (SyncBlockProver): %d\n", lastGeneratedBlockNumber)
// 	fmt.Printf("lastPostedBlockNumber (SyncBlockProver): %d\n", lastPostedBlockNumber)
// 	for blockNumber := lastGeneratedBlockNumber + 1; blockNumber <= lastPostedBlockNumber; blockNumber++ {
// 		// validityWitnessBlockNumber := p.blockBuilder.LastWitnessGeneratedBlockNumber()
// 		fmt.Printf("IMPORTANT: Block %d proof is processing\n", blockNumber)
// 		if err = p.generateValidityProof(blockNumber); err != nil {
// 			return err
// 		}
// 		fmt.Printf("Block %d is reflected\n", blockNumber)
// 	}

// 	return nil
// }

func (p *blockValidityProver) generateValidityProof(blockNumber uint32) error {
	lastValidityProof, err := p.blockBuilder.ValidityProofByBlockNumber(blockNumber - 1)
	if err != nil && err.Error() != ErrGenesisValidityProof.Error() {
		// if err.Error() != ErrGenesisValidityProof.Error() {
		//  var ErrLastValidityProofFail = errors.New("last validity proof fail")
		// 	return errors.Join(ErrLastValidityProofFail, err)
		// }

		if err.Error() != ErrNoValidityProofByBlockNumber.Error() {
			return ErrNoValidityProofByBlockNumber
		}
	}
	fmt.Printf("====== generateValidityProof: block %d ========\n", blockNumber)

	validityWitness, err := p.blockBuilder.UpdateValidityWitnessByBlockNumber(blockNumber)
	if err != nil {
		fmt.Printf("WARNING: failed to update validity witness (generateValidityProof): %v\n", err)
		if errors.Is(err, ErrBlockContentByBlockNumber) {
			return ErrBlockContentByBlockNumber
		}

		if errors.Is(err, ErrRootBlockNumberNotFound) {
			return ErrRootBlockNumberNotFound
		}

		panic(fmt.Errorf("last validity witness error: %w", err))
	}
	fmt.Printf("SenderFlag: %v\n", validityWitness.BlockWitness.Signature.SenderFlag)
	fmt.Printf("validityWitness.BlockWitness.Block.BlockNumber: %d\n", validityWitness.BlockWitness.Block.BlockNumber)

	// encodedBlockWitness, err := json.Marshal(validityWitness.BlockWitness)
	// if err != nil {
	// 	panic("marshal validity witness error")
	// }
	// fmt.Printf("encodedBlockWitness (SyncBlockProver): %s\n", encodedBlockWitness)

	fmt.Printf("validityWitness AccountRegistrationProofs: %v\n", validityWitness.ValidityTransitionWitness.AccountRegistrationProofs)

	// let prev_validity_proof = if let Some(req_prev_validity_proof) = &req.prev_validity_proof {
	//     log::debug!("requested proof size: {}", req_prev_validity_proof.len());
	//     let prev_validity_proof =
	//         decode_plonky2_proof(req_prev_validity_proof, &validity_circuit_data)
	//             .map_err(error::ErrorInternalServerError)?;
	//     validity_circuit_data
	//         .verify(prev_validity_proof.clone())
	//         .map_err(error::ErrorInternalServerError)?;

	//     Some(prev_validity_proof)
	// } else {
	//     None
	// };

	// let prev_pis = if prev_validity_proof.is_some() {
	//     ValidityPublicInputs::from_pis(&prev_validity_proof.as_ref().unwrap().public_inputs)
	// } else {
	//     ValidityPublicInputs::genesis()
	// };
	// if prev_pis.public_state.account_tree_root != validity_witness.block_witness.prev_account_tree_root {
	//     let response = ProofResponse {
	//         success: false,
	//         proof: None,
	//         error_message: Some("account tree root is mismatch".to_string()),
	//     };
	//     println!("block tree root is mismatch: {} != {}", prev_pis.public_state.account_tree_root, validity_witness.block_witness.prev_account_tree_root);
	//     return Ok(HttpResponse::Ok().json(response));
	// }
	// if prev_pis.public_state.block_tree_root != validity_witness.block_witness.prev_block_tree_root {
	//     let response = ProofResponse {
	//         success: false,
	//         proof: None,
	//         error_message: Some("block tree root is mismatch".to_string()),
	//     };
	//     println!("block tree root is mismatch: {} != {}", prev_pis.public_state.block_tree_root, validity_witness.block_witness.prev_block_tree_root);
	//     return Ok(HttpResponse::Ok().json(response));
	// }

	// validation
	var prevValidityPublicInputs *ValidityPublicInputs
	if lastValidityProof == nil {
		prevValidityPublicInputs = new(ValidityPublicInputs).Genesis()
	} else {
		prevBalanceProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*lastValidityProof)
		if err != nil {
			fmt.Printf("WARNING: failed to generate validity proof (generateValidityProof): %v\n", err)
			return err
		}
		prevValidityPublicInputs = new(ValidityPublicInputs).FromPublicInputs(prevBalanceProofWithPis.PublicInputs)
	}

	for targetBlockNumber, merkleTrees := range p.blockBuilder.MerkleTreeHistory.MerkleTrees {
		fmt.Printf("merkleTrees[%d].accountTreeRoot: %v\n", targetBlockNumber, merkleTrees.AccountTree.GetRoot().String())
	}
	if !prevValidityPublicInputs.PublicState.AccountTreeRoot.Equal(validityWitness.BlockWitness.PrevAccountTreeRoot) {
		fmt.Printf("prevValidityPublicInputs.PublicState.BlockNumber: %d\n", prevValidityPublicInputs.PublicState.BlockNumber)
		fmt.Printf("validityWitness.BlockWitness.Block.BlockNumber: %d\n", validityWitness.BlockWitness.Block.BlockNumber)
		fmt.Printf(
			"prev account tree root is mismatch: %s != %s\n",
			prevValidityPublicInputs.PublicState.AccountTreeRoot.String(),
			validityWitness.BlockWitness.PrevAccountTreeRoot.String(),
		)
		return errors.New("prev account tree root is mismatch")
	}

	if !prevValidityPublicInputs.PublicState.BlockTreeRoot.Equal(validityWitness.BlockWitness.PrevBlockTreeRoot) {
		fmt.Printf(
			"prev block tree root is mismatch: %s != %s\n",
			prevValidityPublicInputs.PublicState.BlockTreeRoot.String(),
			validityWitness.BlockWitness.PrevBlockTreeRoot.String(),
		)
		return errors.New("prev block tree root is mismatch")
	}

	validityProof, err := p.requestAndFetchBlockValidityProof(validityWitness, lastValidityProof)
	if err != nil {
		fmt.Printf("WARNING: failed to generate validity proof: %v\n", err)
		return errors.Join(ErrRequestAndFetchBlockValidityProofFail, err)
	}
	fmt.Printf("finished validity proof generation\n")

	validityProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(validityProof)
	if err != nil {
		var ErrNewCompressedPlonky2ProofFromBase64StringFail = errors.New("new compressed plonky2 proof from base64 string fail")
		return errors.Join(ErrNewCompressedPlonky2ProofFromBase64StringFail, err)
	}
	validityPubicInputs := new(ValidityPublicInputs).FromPublicInputs(validityProofWithPis.PublicInputs)
	fmt.Printf("SyncBlockProver block_proof block number: %d\n", validityPubicInputs.PublicState.BlockNumber)
	fmt.Printf("SyncBlockProver block_proof prev account tree root: %s\n", validityPubicInputs.PublicState.PrevAccountTreeRoot.String())
	fmt.Printf("SyncBlockProver block_proof account tree root: %s\n", validityPubicInputs.PublicState.AccountTreeRoot.String())

	err = p.blockBuilder.SetValidityProof(validityWitness.BlockWitness.Block.Hash(), validityProof)
	if err != nil {
		panic(err)
	}

	return nil
}

func (p *blockValidityProver) requestAndFetchBlockValidityProof(validityWitness *ValidityWitness, lastValidityProof *string) (validityProof string, err error) {
	blockHash := validityWitness.BlockWitness.Block.Hash()
	blockNumber := validityWitness.BlockWitness.Block.BlockNumber
	fmt.Printf("Block %d is requested: %s\n", blockNumber, blockHash.String())

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
