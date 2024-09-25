package types

import (
	"encoding/json"
	"intmax2-node/internal/block_post_service"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
)

type ValidityTransitionWitness struct {
	SenderLeaves              []SenderLeaf                   `json:"senderLeaves"`
	BlockMerkleProof          intMaxTree.PoseidonMerkleProof `json:"blockMerkleProof"`
	AccountRegistrationProofs AccountRegistrationProofs      `json:"accountRegistrationProofs"`
	AccountUpdateProofs       AccountUpdateProofs            `json:"accountUpdateProofs"`
}

func (vtw *ValidityTransitionWitness) MarshalJSON() ([]byte, error) {
	result := ValidityTransitionWitnessFlatten{
		SenderLeaves:              vtw.SenderLeaves,
		BlockMerkleProof:          vtw.BlockMerkleProof,
		AccountRegistrationProofs: make([]intMaxTree.IndexedInsertionProof, 0),
		AccountUpdateProofs:       make([]intMaxTree.IndexedUpdateProof, 0),
	}

	if vtw.AccountRegistrationProofs.IsValid {
		result.AccountRegistrationProofs = vtw.AccountRegistrationProofs.Proofs
	}

	if vtw.AccountUpdateProofs.IsValid {
		result.AccountUpdateProofs = vtw.AccountUpdateProofs.Proofs
	}

	return json.Marshal(&result)
}

func (vtw *ValidityTransitionWitness) Set(other *ValidityTransitionWitness) *ValidityTransitionWitness {
	vtw.SenderLeaves = make([]SenderLeaf, len(other.SenderLeaves))
	copy(vtw.SenderLeaves, other.SenderLeaves)
	vtw.BlockMerkleProof.Set(&other.BlockMerkleProof)
	vtw.AccountRegistrationProofs.Set(&other.AccountRegistrationProofs)
	vtw.AccountUpdateProofs.Set(&other.AccountUpdateProofs)

	return vtw
}

func (vtw *ValidityTransitionWitness) Genesis() *ValidityTransitionWitness {
	senderLeaves := make([]SenderLeaf, 0)
	accountRegistrationProofs := make([]intMaxTree.IndexedInsertionProof, 0)
	accountUpdateProofs := make([]intMaxTree.IndexedUpdateProof, 0)

	// Create a empty block hash tree
	blockHashTree, err := intMaxTree.NewBlockHashTreeWithInitialLeaves(intMaxTree.BLOCK_HASH_TREE_HEIGHT, nil)
	if err != nil {
		panic(err)
	}

	prevRoot := blockHashTree.GetRoot()
	prevLeafHash := new(intMaxTree.BlockHashLeaf).SetDefault().Hash()
	blockMerkleProof, _, err := blockHashTree.Prove(0)
	if err != nil {
		panic(err)
	}

	// verify
	err = blockMerkleProof.Verify(prevLeafHash, 0, prevRoot)
	if err != nil {
		panic(err)
	}

	genesisBlock := new(block_post_service.PostedBlock).Genesis()
	genesisBlockHash := intMaxTree.NewBlockHashLeaf(genesisBlock.Hash())
	newRoot, err := blockHashTree.AddLeaf(0, genesisBlockHash)
	if err != nil {
		panic(err)
	}
	err = blockMerkleProof.Verify(genesisBlockHash.Hash(), 0, newRoot)
	if err != nil {
		panic(err)
	}

	return &ValidityTransitionWitness{
		SenderLeaves:     senderLeaves,
		BlockMerkleProof: blockMerkleProof,
		AccountRegistrationProofs: AccountRegistrationProofs{
			IsValid: false,
			Proofs:  accountRegistrationProofs,
		},
		AccountUpdateProofs: AccountUpdateProofs{
			IsValid: false,
			Proofs:  accountUpdateProofs,
		},
	}
}

func (vtw *ValidityTransitionWitness) Compress(maxAccountID uint64) (compressed *CompressedValidityTransitionWitness, err error) {
	compressed = &CompressedValidityTransitionWitness{
		SenderLeaves:             vtw.SenderLeaves,
		BlockMerkleProof:         vtw.BlockMerkleProof,
		CommonAccountMerkleProof: make([]*intMaxGP.PoseidonHashOut, 0),
	}

	significantHeight := int(EffectiveBits(uint(maxAccountID)))

	if vtw.AccountRegistrationProofs.IsValid {
		accountRegistrationProofs := vtw.AccountRegistrationProofs.Proofs
		compressed.CommonAccountMerkleProof = accountRegistrationProofs[0].LowLeafProof.Siblings[significantHeight:]
		significantAccountRegistrationProofs := make([]AccountRegistrationProofOrDummy, 0)
		for _, proof := range accountRegistrationProofs {
			var lowLeafProof *intMaxTree.PoseidonMerkleProof = nil
			if !proof.LowLeafProof.IsDummy(intMaxTree.ACCOUNT_TREE_HEIGHT) {
				for i := 0; i < int(intMaxTree.ACCOUNT_TREE_HEIGHT)-significantHeight; i++ {
					if !proof.LowLeafProof.Siblings[significantHeight+i].Equal(compressed.CommonAccountMerkleProof[i]) {
						panic("invalid low leaf proof")
					}

					lowLeafProof = &intMaxTree.PoseidonMerkleProof{
						Siblings: proof.LowLeafProof.Siblings[:significantHeight],
					}
				}
			}

			var leafProof *intMaxTree.PoseidonMerkleProof = nil
			if !proof.LeafProof.IsDummy(intMaxTree.ACCOUNT_TREE_HEIGHT) {
				for i := 0; i < int(intMaxTree.ACCOUNT_TREE_HEIGHT)-significantHeight; i++ {
					if !proof.LeafProof.Siblings[significantHeight+i].Equal(compressed.CommonAccountMerkleProof[i]) {
						panic("invalid leaf proof")
					}

					leafProof = &intMaxTree.PoseidonMerkleProof{
						Siblings: proof.LeafProof.Siblings[:significantHeight],
					}
				}
			}

			significantAccountRegistrationProofs = append(significantAccountRegistrationProofs, AccountRegistrationProofOrDummy{
				LowLeafProof: lowLeafProof,
				LeafProof:    leafProof,
				Index:        uint64(proof.Index),
				LowLeafIndex: uint64(proof.LowLeafIndex),
				PrevLowLeaf:  *proof.PrevLowLeaf,
			})
		}

		compressed.SignificantAccountRegistrationProofs = &significantAccountRegistrationProofs
	}

	if vtw.AccountUpdateProofs.IsValid {
		accountUpdateProofs := vtw.AccountUpdateProofs.Proofs
		compressed.CommonAccountMerkleProof = accountUpdateProofs[0].LeafProof.Siblings[significantHeight:]
		significantAccountUpdateProofs := make([]intMaxTree.IndexedUpdateProof, 0)
		for _, proof := range accountUpdateProofs {
			for i := 0; i < int(intMaxTree.ACCOUNT_TREE_HEIGHT)-significantHeight; i++ {
				if proof.LeafProof.Siblings[significantHeight+i].Equal(compressed.CommonAccountMerkleProof[i]) {
					panic("invalid leaf proof")
				}

				significantAccountUpdateProofs = append(significantAccountUpdateProofs, intMaxTree.IndexedUpdateProof{
					LeafProof: intMaxTree.IndexedMerkleProof{
						Siblings: proof.LeafProof.Siblings[:significantHeight],
					},
					LeafIndex: proof.LeafIndex,
					PrevLeaf:  proof.PrevLeaf,
				})
			}
		}

		compressed.SignificantAccountUpdateProofs = &significantAccountUpdateProofs
	}

	return compressed, nil
}
