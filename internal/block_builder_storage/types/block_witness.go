package types

import (
	"encoding/json"
	"fmt"
	"intmax2-node/internal/block_post_service"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
)

type BlockWitness struct {
	Block                   *block_post_service.PostedBlock      `json:"block"`
	Signature               SignatureContent                     `json:"signature"`
	PublicKeys              []intMaxTypes.Uint256                `json:"pubkeys"`
	PrevAccountTreeRoot     *intMaxTree.PoseidonHashOut          `json:"prevAccountTreeRoot"`
	PrevBlockTreeRoot       *intMaxTree.PoseidonHashOut          `json:"prevBlockTreeRoot"`
	AccountIdPacked         *AccountIdPacked                     `json:"accountIdPacked,omitempty"` // in account id case
	AccountMerkleProofs     *[]AccountMerkleProof                `json:"accountMerkleProofs"`       // in account id case
	AccountMembershipProofs *[]intMaxTree.IndexedMembershipProof `json:"accountMembershipProofs"`   // in pubkey case
}

func (bw *BlockWitness) MarshalJSON() ([]byte, error) {
	result := BlockWitnessFlatten{
		Block:                   bw.Block,
		Signature:               bw.Signature,
		PublicKeys:              bw.PublicKeys,
		PrevAccountTreeRoot:     bw.PrevAccountTreeRoot,
		PrevBlockTreeRoot:       bw.PrevBlockTreeRoot,
		AccountIdPacked:         bw.AccountIdPacked,
		AccountMerkleProofs:     make([]AccountMerkleProof, 0),
		AccountMembershipProofs: make([]intMaxTree.IndexedMembershipProof, 0),
	}

	if bw.AccountMembershipProofs != nil {
		result.AccountMembershipProofs = *bw.AccountMembershipProofs
	}
	if bw.AccountMerkleProofs != nil {
		result.AccountMerkleProofs = *bw.AccountMerkleProofs
	}

	return json.Marshal(&result)
}

func (bw *BlockWitness) Set(blockWitness *BlockWitness) *BlockWitness {
	bw.Block = new(block_post_service.PostedBlock).Set(blockWitness.Block)
	bw.Signature.Set(&blockWitness.Signature)
	bw.PublicKeys = make([]intMaxTypes.Uint256, len(blockWitness.PublicKeys))
	copy(bw.PublicKeys, blockWitness.PublicKeys)

	bw.PrevAccountTreeRoot = new(intMaxGP.PoseidonHashOut).Set(blockWitness.PrevAccountTreeRoot)
	bw.PrevBlockTreeRoot = new(intMaxGP.PoseidonHashOut).Set(blockWitness.PrevBlockTreeRoot)
	bw.AccountIdPacked = new(AccountIdPacked).Set(blockWitness.AccountIdPacked)
	if blockWitness.AccountMerkleProofs != nil {
		accountMerkleProofs := make([]AccountMerkleProof, len(*blockWitness.AccountMerkleProofs))
		copy(accountMerkleProofs, *blockWitness.AccountMerkleProofs)
		bw.AccountMerkleProofs = &accountMerkleProofs
	}
	if blockWitness.AccountMembershipProofs != nil {
		accountMembershipProofs := make([]intMaxTree.IndexedMembershipProof, len(*blockWitness.AccountMembershipProofs))
		copy(accountMembershipProofs, *blockWitness.AccountMembershipProofs)
		bw.AccountMembershipProofs = &accountMembershipProofs
	}

	return bw
}

func (bw *BlockWitness) Genesis() *BlockWitness {
	blockHashTree, err := intMaxTree.NewBlockHashTreeWithInitialLeaves(intMaxTree.BLOCK_HASH_TREE_HEIGHT, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Genesis blockHashTree leaves: %v\n", blockHashTree.Leaves)
	prevBlockTreeRoot, _, _ := blockHashTree.GetCurrentRootCountAndSiblings()
	accountTree, err := intMaxTree.NewAccountTree(intMaxTree.ACCOUNT_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}
	prevAccountTreeRoot := accountTree.GetRoot()

	return &BlockWitness{
		Block:                   new(block_post_service.PostedBlock).Genesis(),
		Signature:               SignatureContent{},
		PublicKeys:              make([]intMaxTypes.Uint256, 0),
		PrevAccountTreeRoot:     prevAccountTreeRoot,
		PrevBlockTreeRoot:       &prevBlockTreeRoot,
		AccountIdPacked:         nil,
		AccountMerkleProofs:     nil,
		AccountMembershipProofs: nil,
	}
}

func (bw *BlockWitness) Compress(maxAccountID uint64) (compressed *CompressedBlockWitness, err error) {
	compressed = &CompressedBlockWitness{
		Block:                    bw.Block,
		Signature:                bw.Signature,
		PublicKeys:               bw.PublicKeys,
		PrevAccountTreeRoot:      *bw.PrevAccountTreeRoot,
		PrevBlockTreeRoot:        *bw.PrevBlockTreeRoot,
		AccountIdPacked:          bw.AccountIdPacked,
		CommonAccountMerkleProof: make([]*intMaxGP.PoseidonHashOut, 0),
	}

	significantHeight := EffectiveBits(uint(maxAccountID))

	if bw.AccountMerkleProofs != nil {
		if len(*bw.AccountMerkleProofs) == 0 {
			significantAccountMerkleProofs := make([]AccountMerkleProof, 0)
			compressed.SignificantAccountMerkleProofs = &significantAccountMerkleProofs
		} else {
			accountMerkleProofs := *bw.AccountMerkleProofs
			compressed.CommonAccountMerkleProof = accountMerkleProofs[0].MerkleProof.Siblings[significantHeight:]
			significantAccountMerkleProofs := make([]AccountMerkleProof, 0)
			for _, proof := range accountMerkleProofs {
				for i := 0; i < int(intMaxTree.ACCOUNT_TREE_HEIGHT)-int(significantHeight); i++ {
					if !proof.MerkleProof.Siblings[int(significantHeight)+i].Equal(compressed.CommonAccountMerkleProof[i]) {
						panic("invalid common account merkle proof")
					}
				}

				significantMerkleProof := proof.MerkleProof.Siblings[:significantHeight]
				significantAccountMerkleProofs = append(significantAccountMerkleProofs, AccountMerkleProof{
					MerkleProof: intMaxTree.IndexedMerkleProof(
						intMaxTree.PoseidonMerkleProof{
							Siblings: significantMerkleProof,
						},
					),
					Leaf: proof.Leaf,
				})
			}

			compressed.SignificantAccountMerkleProofs = &significantAccountMerkleProofs
		}
	}

	if bw.AccountMembershipProofs != nil {
		if len(*bw.AccountMembershipProofs) == 0 {
			significantAccountMembershipProofs := make([]intMaxTree.IndexedMembershipProof, 0)
			compressed.SignificantAccountMembershipProofs = &significantAccountMembershipProofs
		} else {
			accountMembershipProofs := *bw.AccountMembershipProofs
			compressed.CommonAccountMerkleProof = accountMembershipProofs[0].LeafProof.Siblings[significantHeight:]
			significantAccountMembershipProofs := make([]intMaxTree.IndexedMembershipProof, 0)
			for _, proof := range accountMembershipProofs {
				for i := 0; i < int(intMaxTree.ACCOUNT_TREE_HEIGHT)-int(significantHeight); i++ {
					if !proof.LeafProof.Siblings[int(significantHeight)+i].Equal(compressed.CommonAccountMerkleProof[i]) {
						panic("invalid common account merkle proof")
					}
				}

				significantMerkleProof := proof.LeafProof.Siblings[:significantHeight]
				significantAccountMembershipProofs = append(significantAccountMembershipProofs, intMaxTree.IndexedMembershipProof{
					LeafProof: intMaxTree.IndexedMerkleProof{
						Siblings: significantMerkleProof,
					},
					LeafIndex:  proof.LeafIndex,
					Leaf:       proof.Leaf,
					IsIncluded: proof.IsIncluded,
				})
			}

			compressed.SignificantAccountMembershipProofs = &significantAccountMembershipProofs
		}
	}

	return compressed, nil
}

func (bw *BlockWitness) MainValidationPublicInputs() *MainValidationPublicInputs {
	if new(block_post_service.PostedBlock).Genesis().Equals(bw.Block) {
		validityPis := new(ValidityPublicInputs).Genesis()
		return &MainValidationPublicInputs{
			PrevBlockHash:       new(block_post_service.PostedBlock).Genesis().PrevBlockHash,
			BlockHash:           validityPis.PublicState.BlockHash,
			DepositTreeRoot:     validityPis.PublicState.DepositTreeRoot,
			AccountTreeRoot:     validityPis.PublicState.AccountTreeRoot,
			TxTreeRoot:          validityPis.TxTreeRoot,
			SenderTreeRoot:      validityPis.SenderTreeRoot,
			BlockNumber:         validityPis.PublicState.BlockNumber,
			IsRegistrationBlock: false, // genesis block is not a registration block
			IsValid:             validityPis.IsValidBlock,
		}
	}

	result := true
	block := new(block_post_service.PostedBlock).Set(bw.Block)
	signature := new(SignatureContent).Set(&bw.Signature)
	publicKeys := make([]intMaxTypes.Uint256, len(bw.PublicKeys))
	copy(publicKeys, bw.PublicKeys)

	prevAccountTreeRoot := bw.PrevAccountTreeRoot

	publicKeysHash := GetPublicKeysHash(publicKeys)
	isRegistrationBlock := signature.IsRegistrationBlock
	isPubkeyEq := signature.PublicKeyHash == publicKeysHash
	if isRegistrationBlock {
		if !isPubkeyEq {
			panic("pubkey hash mismatch")
		}
	} else {
		result = result && isPubkeyEq
	}
	if isRegistrationBlock {
		if bw.AccountMembershipProofs == nil {
			panic("account membership proofs should be given")
		}

		// Account exclusion verification
		accountExclusionValue, err := NewAccountExclusionValue(
			prevAccountTreeRoot,
			*bw.AccountMembershipProofs,
			publicKeys,
		)
		if err != nil {
			panic("account exclusion value is invalid: " + err.Error())
		}

		result = result && accountExclusionValue.IsValid
	} else {
		if bw.AccountIdPacked != nil {
			panic("account id packed should be given")
		}

		if bw.AccountMerkleProofs == nil {
			panic("account merkle proofs should be given")
		}

		// Account inclusion verification
		accountInclusionValue, err := NewAccountInclusionValue(
			prevAccountTreeRoot,
			bw.AccountIdPacked,
			*bw.AccountMerkleProofs,
			publicKeys,
		)
		if err != nil {
			panic("account inclusion value is invalid: " + err.Error())
		}

		result = result && accountInclusionValue.IsValid
	}

	// Format validation
	formatValidationValue :=
		NewFormatValidationValue(publicKeys, signature)
	result = result && formatValidationValue.IsValid

	if result {
		aggregationValue := NewAggregationValue(publicKeys, signature)
		result = result && aggregationValue.IsValid
	}

	prev_block_hash := block.PrevBlockHash
	blockHash := block.Hash()
	senderTreeRoot := GetSenderTreeRoot(publicKeys, signature.SenderFlag)

	txTreeRoot := signature.TxTreeRoot

	return &MainValidationPublicInputs{
		PrevBlockHash:       prev_block_hash,
		BlockHash:           blockHash,
		DepositTreeRoot:     block.DepositRoot,
		AccountTreeRoot:     prevAccountTreeRoot,
		TxTreeRoot:          txTreeRoot,
		SenderTreeRoot:      senderTreeRoot,
		BlockNumber:         block.BlockNumber,
		IsRegistrationBlock: isRegistrationBlock,
		IsValid:             result,
	}
}
