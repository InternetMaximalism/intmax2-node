package types

import (
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	intMaxTree "intmax2-node/internal/tree"
)

type ValidityWitness struct {
	BlockWitness              *BlockWitness              `json:"blockWitness"`
	ValidityTransitionWitness *ValidityTransitionWitness `json:"validityTransitionWitness"`
}

func (vw *ValidityWitness) Set(validityWitness *ValidityWitness) *ValidityWitness {
	vw.BlockWitness = new(BlockWitness).Set(validityWitness.BlockWitness)
	vw.ValidityTransitionWitness = new(ValidityTransitionWitness).Set(validityWitness.ValidityTransitionWitness)

	return vw
}

func (vw *ValidityWitness) Genesis() *ValidityWitness {
	return &ValidityWitness{
		BlockWitness:              new(BlockWitness).Genesis(),
		ValidityTransitionWitness: new(ValidityTransitionWitness).Genesis(),
	}
}

func (vw *ValidityWitness) Compress(maxAccountID uint64) (*CompressedValidityWitness, error) {
	blockWitness, err := vw.BlockWitness.Compress(maxAccountID)
	if err != nil {
		return nil, err
	}

	validityTransitionWitness, err := vw.ValidityTransitionWitness.Compress(maxAccountID)
	if err != nil {
		return nil, err
	}

	return &CompressedValidityWitness{
		BlockWitness:              blockWitness,
		ValidityTransitionWitness: validityTransitionWitness,
	}, nil
}

func (vw *ValidityWitness) ValidityPublicInputs(log logger.Logger) *ValidityPublicInputs {
	blockWitness := vw.BlockWitness
	validityTransitionWitness := vw.ValidityTransitionWitness

	prevBlockTreeRoot := blockWitness.PrevBlockTreeRoot

	// Check transition block tree root
	block := blockWitness.Block
	defaultLeaf := new(intMaxTree.BlockHashLeaf).SetDefault()
	log.Debugf("old block root: %s", prevBlockTreeRoot.String())
	err := validityTransitionWitness.BlockMerkleProof.Verify(
		defaultLeaf.Hash(),
		int(block.BlockNumber),
		prevBlockTreeRoot,
	)

	if err != nil {
		panic("Block merkle proof is invalid")
	}
	blockHashLeaf := intMaxTree.NewBlockHashLeaf(block.Hash())
	blockTreeRoot := validityTransitionWitness.BlockMerkleProof.GetRoot(blockHashLeaf.Hash(), int(block.BlockNumber))
	log.Debugf("new block root: %s", blockTreeRoot.String())

	mainValidationPis := blockWitness.MainValidationPublicInputs()

	// transition account tree root
	prevAccountTreeRoot := blockWitness.PrevAccountTreeRoot
	accountTreeRoot := new(intMaxGP.PoseidonHashOut).Set(prevAccountTreeRoot)
	log.Debugf("mainValidationPis.IsValid: %v", mainValidationPis.IsValid)
	log.Debugf("mainValidationPis.IsRegistrationBlock: %v", mainValidationPis.IsRegistrationBlock)
	if mainValidationPis.IsValid && mainValidationPis.IsRegistrationBlock {
		accountRegistrationProofs := validityTransitionWitness.AccountRegistrationProofs
		if !accountRegistrationProofs.IsValid {
			panic("account registration proofs should be given")
		}
		for i, senderLeaf := range validityTransitionWitness.SenderLeaves {
			accountRegistrationProof := accountRegistrationProofs.Proofs[i]
			var lastBlockNumber uint32 = 0
			if senderLeaf.IsValid {
				lastBlockNumber = block.BlockNumber
			}

			dummyPublicKey := intMaxAcc.NewDummyPublicKey()
			isDummy := senderLeaf.Sender.Cmp(dummyPublicKey.BigInt()) == 0
			accountTreeRoot, err = accountRegistrationProof.ConditionalGetNewRoot(
				!isDummy,
				senderLeaf.Sender,
				uint64(lastBlockNumber),
				accountTreeRoot,
			)
			if err != nil {
				log.Debugf("senderLeaf.Sender: %s", senderLeaf.Sender.String())
				panic("Invalid account registration proof: " + err.Error())
			}
		}
	}
	if mainValidationPis.IsValid && !mainValidationPis.IsRegistrationBlock {
		accountUpdateProofs := validityTransitionWitness.AccountUpdateProofs
		if !accountUpdateProofs.IsValid {
			panic("account update proofs should be given")
		}
		for i, senderLeaf := range validityTransitionWitness.SenderLeaves {
			accountUpdateProof := accountUpdateProofs.Proofs[i]
			prevLastBlockNumber := uint32(accountUpdateProof.PrevLeaf.Value)
			lastBlockNumber := prevLastBlockNumber
			if senderLeaf.IsValid {
				lastBlockNumber = block.BlockNumber
			}
			accountTreeRoot, err = accountUpdateProof.GetNewRoot(
				senderLeaf.Sender,
				uint64(prevLastBlockNumber),
				uint64(lastBlockNumber),
				accountTreeRoot,
			)

			if err != nil {
				panic("Invalid account update proof")
			}
		}
	}

	log.Debugf("blockNumber (ValidityPublicInputs): %d", block.BlockNumber)
	log.Debugf("prevAccountTreeRoot (ValidityPublicInputs): %s", prevAccountTreeRoot.String())
	log.Debugf("accountTreeRoot (ValidityPublicInputs): %s", accountTreeRoot.String())
	return &ValidityPublicInputs{
		PublicState: &PublicState{
			BlockTreeRoot:       blockTreeRoot,
			PrevAccountTreeRoot: prevAccountTreeRoot,
			AccountTreeRoot:     accountTreeRoot,
			DepositTreeRoot:     block.DepositRoot,
			BlockHash:           mainValidationPis.BlockHash,
			BlockNumber:         block.BlockNumber,
		},
		TxTreeRoot:     mainValidationPis.TxTreeRoot,
		SenderTreeRoot: mainValidationPis.SenderTreeRoot,
		IsValidBlock:   mainValidationPis.IsValid,
	}
}
