package balance_prover_service

import (
	"encoding/hex"
	"errors"
	"intmax2-node/internal/block_validity_prover"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
)

type UserWalletState struct {
	NullifierTree   *intMaxTree.NullifierTree
	NullifierLeaves []intMaxTypes.Bytes32
	AssetTree       *intMaxTree.AssetTree
	Nonce           uint32
	Salt            Salt
	PublicState     *block_validity_prover.PublicState
}

type UserWalletStateInput struct {
	Nullifiers  []string          `json:"nullifiers"`
	Assets      []*AssetLeafInput `json:"assets"`
	Nonce       uint32            `json:"nonce"`
	Salt        SaltInput         `json:"salt"`
	PublicState *PublicStateInput `json:"publicState"`
}

func (input *UserWalletStateInput) FromUserState(value *UserWalletState) *UserWalletStateInput {
	input.Nullifiers = make([]string, len(value.NullifierLeaves))
	for i, leaf := range value.NullifierLeaves {
		input.Nullifiers[i] = hex.EncodeToString(leaf.Bytes())
	}

	input.Assets = make([]*AssetLeafInput, len(value.AssetTree.Leaves))
	for i, leaf := range value.AssetTree.Leaves {
		input.Assets[i] = &AssetLeafInput{
			IsInsufficient: leaf.IsInsufficient,
			Amount:         leaf.Amount.BigInt().String(),
		}
	}

	input.Nonce = value.Nonce
	input.Salt = value.Salt
	input.PublicState = new(PublicStateInput).FromPublicState(value.PublicState)

	return input
}

func (input *UserWalletStateInput) UserState() (*UserWalletState, error) {
	nullifierLeaves := make([]intMaxTypes.Bytes32, len(input.Nullifiers))
	for i, leaf := range input.Nullifiers {
		leafBytes, err := hex.DecodeString(leaf)
		if err != nil {
			return nil, err
		}
		nullifierLeaves[i].FromBytes(leafBytes)
	}

	const base10 = 10
	assetLeaves := make([]*intMaxTree.AssetLeaf, len(input.Assets))
	for i, leaf := range input.Assets {
		amountInt, ok := new(big.Int).SetString(leaf.Amount, base10)
		if !ok {
			return nil, errors.New("invalid amount in UserState")
		}
		amount := new(intMaxTypes.Uint256).FromBigInt(amountInt)

		assetLeaves[i] = &intMaxTree.AssetLeaf{
			IsInsufficient: leaf.IsInsufficient,
			Amount:         amount,
		}
	}

	publicState := input.PublicState.PublicState()

	nullifierTree, err := intMaxTree.NewNullifierTree(intMaxTree.NULLIFIER_TREE_HEIGHT)
	if err != nil {
		return nil, err
	}
	for _, leaf := range nullifierLeaves {
		if _, err := nullifierTree.Insert(leaf); err != nil {
			return nil, err
		}
	}

	zeroAsset := *new(intMaxTree.AssetLeaf).SetDefault()
	assetTree, err := intMaxTree.NewAssetTree(intMaxTree.ASSET_TREE_HEIGHT, assetLeaves, zeroAsset.Hash())
	if err != nil {
		return nil, err
	}
	for _, leaf := range assetLeaves {
		_, nextIndex, _ := assetTree.GetCurrentRootCountAndSiblings()
		if _, err := assetTree.AddLeaf(uint32(nextIndex), leaf); err != nil {
			return nil, err
		}
	}

	return &UserWalletState{
		NullifierTree:   nullifierTree,
		NullifierLeaves: nullifierLeaves,
		AssetTree:       assetTree,
		Nonce:           input.Nonce,
		Salt:            input.Salt,
		PublicState:     publicState,
	}, nil
}
