package block_validity_prover

import (
	"encoding/hex"
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/block_post_service"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/pkg/sql_db/db_app/models"
	"strings"
	"sync"
	"time"

	"github.com/holiman/uint256"
)

const (
	stepDeposits       = "deposits"
	stepBlockContents  = "block-contents"
	stepValidityProofs = "validity-proofs"
)

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

func (p *blockValidityProver) SyncBlockTree(bps BlockSynchronizer, wg *sync.WaitGroup) (err error) {
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()

		var latestSynchronizedDepositIndex uint32
		latestSynchronizedDepositIndex, err = p.FetchLastDepositIndex()
		if err != nil {
			const msg = "failed to fetch last deposit index: %+v"
			p.log.Fatalf(msg, err.Error())
		}

		timeout := 5 * time.Second
		ticker := time.NewTicker(timeout)
		for {
			select {
			case <-p.ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				fmt.Println("balance validity ticker.C")
				err = p.SyncDepositedEvents()
				if err != nil {
					p.log.Fatalf("failed to sync deposited events in balance validity prover: %+v", err)
				}

				err = p.SyncDepositTree(nil, latestSynchronizedDepositIndex+1)
				if err != nil {
					p.log.Fatalf("failed to sync deposit tree in balance validity prover: %+v", err)
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()

		tickerEventWatcher := time.NewTicker(p.cfg.BlockValidityProver.TimeoutForEventWatcher)
		for {
			select {
			case <-p.ctx.Done():
				tickerEventWatcher.Stop()
				return
			case <-tickerEventWatcher.C:
				fmt.Println("block content ticker.C")
				// sync block content
				var startBlock uint64
				startBlock, err := p.LastSeenBlockPostedEventBlockNumber()
				if err != nil {
					startBlock = p.cfg.Blockchain.RollupContractDeployedBlockNumber
				}
				fmt.Printf("startBlock of LastSeenBlockPostedEventBlockNumber: %d\n", startBlock)

				var endBlock uint64
				endBlock, err = p.syncBlockContent(bps, startBlock)
				if err != nil {
					panic(err)
				}
				fmt.Printf("endBlock of LastSeenBlockPostedEventBlockNumber: %d\n", endBlock)

				err = p.SetLastSeenBlockPostedEventBlockNumber(endBlock)
				if err != nil {
					var ErrSetLastSeenBlockPostedEventBlockNumberFail = errors.New("set last seen block posted event block number fail")
					panic(errors.Join(ErrSetLastSeenBlockPostedEventBlockNumberFail, err))
				}

				fmt.Printf("Block %d is searched\n", endBlock)
			}
		}
	}()

	// wg.Add(1)
	// go func() {
	// 	defer func() {
	// 		wg.Done()
	// 	}()

	// 	err = p.SyncBlockValidityWitness()
	// 	if err != nil {
	// 		var ErrSyncBlockProverWithBlockNumberFail = errors.New("failed to sync block validity witness")
	// 		panic(errors.Join(ErrSyncBlockProverWithBlockNumberFail, err))
	// 	}
	// }()

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()

		err = p.syncBlockValidityProof()
		if err != nil {
			var ErrSyncBlockValidityProofFail = errors.New("failed to sync block validity proof")
			panic(errors.Join(ErrSyncBlockValidityProofFail, err))
		}
	}()

	return nil
}

func (p *blockValidityProver) SyncBlockTreeStep(bps BlockSynchronizer, step string) (err error) {
	syncDeposits := func() error {
		var latestSynchronizedDepositIndex uint32
		latestSynchronizedDepositIndex, err = p.FetchLastDepositIndex()
		if err != nil {
			return fmt.Errorf("failed to fetch last deposit index: %+v", err.Error())
		}

		err = p.SyncDepositedEvents()
		if err != nil {
			return fmt.Errorf("failed to sync deposited events: %+v", err.Error())
		}

		err = p.SyncDepositTree(nil, latestSynchronizedDepositIndex+1)
		if err != nil {
			return fmt.Errorf("failed to sync deposit tree: %w", err)
		}

		return nil
	}

	syncBlockContent := func() error {
		// sync block content
		var startBlock uint64
		startBlock, err := p.LastSeenBlockPostedEventBlockNumber()
		if err != nil {
			startBlock = p.cfg.Blockchain.RollupContractDeployedBlockNumber
		}
		fmt.Printf("startBlock of LastSeenBlockPostedEventBlockNumber: %d\n", startBlock)

		var endBlock uint64
		endBlock, err = p.syncBlockContent(bps, startBlock)
		if err != nil {
			return nil
		}
		fmt.Printf("endBlock of LastSeenBlockPostedEventBlockNumber: %d\n", endBlock)

		err = p.SetLastSeenBlockPostedEventBlockNumber(endBlock)
		if err != nil {
			var ErrSetLastSeenBlockPostedEventBlockNumberFail = errors.New("set last seen block posted event block number fail")
			return errors.Join(ErrSetLastSeenBlockPostedEventBlockNumberFail, err)
		}

		fmt.Printf("Block %d is searched\n", endBlock)

		return nil
	}

	syncValidityProof := func() error {
		lastGeneratedProofBlockNumber, err := p.blockBuilder.LastGeneratedProofBlockNumber()
		if err != nil {
			var ErrLastGeneratedProofBlockNumberFail = errors.New("last generated proof block number fail")
			return errors.Join(ErrLastGeneratedProofBlockNumberFail, err)
		}

		if err := p.generateValidityProof(lastGeneratedProofBlockNumber + 1); err != nil {
			if errors.Is(err, ErrBlockContentByBlockNumber) {
				fmt.Printf("WARNING: block content %d is not found\n", lastGeneratedProofBlockNumber+1)
				return nil
			}

			if errors.Is(err, ErrRootBlockNumberNotFound) {
				fmt.Printf("WARNING: root block number %d not found\n", lastGeneratedProofBlockNumber+1)
				return nil
			}

			return err
		}

		fmt.Printf("Block %d is done (syncBlockValidityProof)\n", lastGeneratedProofBlockNumber+1)

		return nil
	}

	switch step {
	case stepDeposits:
		err = syncDeposits()
		if err != nil {
			return fmt.Errorf("failed to sync deposits: %v", err)
		}
	case stepBlockContents:
		err = syncBlockContent()
		if err != nil {
			return fmt.Errorf("failed to sync block content: %v", err)
		}
	case stepValidityProofs:
		err = syncValidityProof()
		if err != nil {
			return fmt.Errorf("failed to sync block validity proof: %v", err)
		}
	default:
		stepNames := []string{stepDeposits, stepBlockContents, stepValidityProofs}
		return fmt.Errorf("step must be one of %s", strings.Join(stepNames, ", "))
	}

	return nil
}

func (p *blockValidityProver) syncBlockContent(bps BlockSynchronizer, startBlock uint64) (lastEventSeenBlockNumber uint64, err error) {
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

	var events []*bindings.RollupBlockPosted
	events, _, err = bps.FetchNewPostedBlocks(startBlock, &endBlock)
	if err != nil {
		return startBlock, errors.Join(ErrFetchNewPostedBlocksFail, err)
	}

	if len(events) == 0 {
		fmt.Printf("Scroll Block %d is synchronized (SyncBlockTree)\n", endBlock)
		return endBlock, nil
	}

	timeout := 1 * time.Second
	tickerEventWatcher := time.NewTicker(timeout)
	defer func() {
		if tickerEventWatcher != nil {
			tickerEventWatcher.Stop()
		}
	}()

	nextBN := startBlock
	for key := range events {
		select {
		case <-p.ctx.Done():
			p.log.Warnf("Received cancel signal from context, stopping...")
			return startBlock, p.ctx.Err()
		case <-tickerEventWatcher.C:
			fmt.Println("tickerEventWatcher.C")
			_, err := syncBlockContentWithEvent(p.blockBuilder, bps, p.log, events[key])
			if err != nil {
				return nextBN, errors.Join(ErrProcessingBlocksFail, err)
			}
			nextBN = events[key].Raw.BlockNumber
		}
	}

	return nextBN, nil
}

func syncBlockContentWithEvent(
	blockBuilder *mockBlockBuilder,
	bps BlockSynchronizer,
	log logger.Logger,
	event *bindings.RollupBlockPosted,
) (*models.BlockContentWithProof, error) {
	intMaxBlockNumber := uint32(event.BlockNumber.Uint64())
	newBlockContent, err := blockBuilder.db.BlockContentByBlockNumber(intMaxBlockNumber)
	if err == nil {
		return newBlockContent, nil
	}

	if err.Error() != "not found" {
		return nil, errors.Join(ErrProcessingBlocksFail, err)
	}

	var cd []byte
	cd, err = bps.FetchScrollCalldataByHash(event.Raw.TxHash)
	if err != nil {
		return nil, errors.Join(ErrFetchScrollCalldataByHashFail, err)
	}

	postedBlock := block_post_service.NewPostedBlock(
		event.PrevBlockHash,
		event.DepositTreeRoot,
		intMaxBlockNumber,
		event.SignatureHash,
	)

	// Update account tree
	var blockContent *intMaxTypes.BlockContent
	blockContent, err = FetchIntMaxBlockContentByCalldata(cd, postedBlock, blockBuilder)
	if err != nil {
		err = errors.Join(ErrFetchIntMaxBlockContentByCalldataFail, err)
		switch {
		case errors.Is(err, ErrUnknownAccountID):
			const msg = "block %d is ErrUnknownAccountID"
			log.WithError(err).Errorf(msg, intMaxBlockNumber)
		case errors.Is(err, ErrCannotDecodeAddress):
			const msg = "block %d is ErrCannotDecodeAddress"
			log.WithError(err).Errorf(msg, intMaxBlockNumber)
		default:
			const msg = "block %d processing error occurred"
			log.WithError(err).Errorf(msg, intMaxBlockNumber)
		}

		const msg = "processing of block %d error occurred"
		log.Debugf(msg, intMaxBlockNumber)
	} else {
		const msg = "block %d is found (SyncBlockTree, Scroll block number: %d)"
		log.Debugf(msg, intMaxBlockNumber, event.Raw.BlockNumber)
	}

	senders := make([]intMaxTypes.ColumnSender, len(blockContent.Senders))
	for i, sender := range blockContent.Senders {
		senders[i] = intMaxTypes.ColumnSender{
			PublicKey: hex.EncodeToString(sender.PublicKey.ToAddress().Bytes()),
			AccountID: sender.AccountID,
			IsSigned:  sender.IsSigned,
		}
	}

	newBlockContent, err = blockBuilder.CreateBlockContent(postedBlock, blockContent)
	if err != nil {
		return nil, errors.Join(ErrCreateBlockContentFail, err)
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

func (p *blockValidityProver) syncBlockValidityProof() error {
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
