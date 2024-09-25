package types

import (
	"encoding/binary"
	"errors"
	"intmax2-node/internal/block_post_service"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/ffg"
)

const NumPublicStateBytes = Int32Key*Int5Key + Int4Key

type PublicState struct {
	BlockTreeRoot       *intMaxGP.PoseidonHashOut `json:"blockTreeRoot"`
	PrevAccountTreeRoot *intMaxGP.PoseidonHashOut `json:"prevAccountTreeRoot"`
	AccountTreeRoot     *intMaxGP.PoseidonHashOut `json:"accountTreeRoot"`
	DepositTreeRoot     common.Hash               `json:"depositTreeRoot"`
	BlockHash           common.Hash               `json:"blockHash"`
	BlockNumber         uint32                    `json:"blockNumber"`
}

func (ps *PublicState) Genesis() *PublicState {
	blockTree, err := NewBlockHashTree(intMaxTree.BLOCK_HASH_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}

	genesisBlockHash := new(block_post_service.PostedBlock).Genesis().Hash()
	blockTreeRoot := blockTree.GetRoot()

	accountTree, err := intMaxTree.NewAccountTree(intMaxTree.ACCOUNT_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("genesis accountTreeRoot: %s\n", accountTree.GetRoot().String())
	depositTree, err := intMaxTree.NewDepositTree(intMaxTree.DEPOSIT_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}

	prevAccountTreeRoot := accountTree.GetRoot()
	accountTreeRoot := accountTree.GetRoot()
	depositTreeRoot, _, _ := depositTree.GetCurrentRootCountAndSiblings()
	return &PublicState{
		BlockTreeRoot:       blockTreeRoot,
		PrevAccountTreeRoot: prevAccountTreeRoot,
		AccountTreeRoot:     accountTreeRoot,
		DepositTreeRoot:     depositTreeRoot,
		BlockHash:           genesisBlockHash,
		BlockNumber:         1,
	}
}

func (ps *PublicState) Set(other *PublicState) *PublicState {
	if other == nil {
		ps = nil
		return nil
	}

	ps.BlockTreeRoot = new(intMaxGP.PoseidonHashOut).Set(other.BlockTreeRoot)
	ps.PrevAccountTreeRoot = new(intMaxGP.PoseidonHashOut).Set(other.PrevAccountTreeRoot)
	ps.AccountTreeRoot = new(intMaxGP.PoseidonHashOut).Set(other.AccountTreeRoot)
	ps.DepositTreeRoot = other.DepositTreeRoot
	ps.BlockHash = other.BlockHash
	ps.BlockNumber = other.BlockNumber

	return ps
}

func (ps *PublicState) Equal(other *PublicState) bool {
	if !ps.BlockTreeRoot.Equal(other.BlockTreeRoot) {
		return false
	}
	if !ps.PrevAccountTreeRoot.Equal(other.PrevAccountTreeRoot) {
		return false
	}
	if !ps.AccountTreeRoot.Equal(other.AccountTreeRoot) {
		return false
	}
	if ps.DepositTreeRoot != other.DepositTreeRoot {
		return false
	}
	if ps.BlockHash != other.BlockHash {
		return false
	}
	if ps.BlockNumber != other.BlockNumber {
		return false
	}

	return true
}

func (ps *PublicState) Marshal() []byte {
	buf := make([]byte, NumPublicStateBytes)
	offset := 0

	copy(buf[offset:offset+Int32Key], ps.BlockTreeRoot.Marshal())
	offset += Int32Key

	copy(buf[offset:offset+Int32Key], ps.PrevAccountTreeRoot.Marshal())
	offset += Int32Key

	copy(buf[offset:offset+Int32Key], ps.AccountTreeRoot.Marshal())
	offset += Int32Key

	copy(buf[offset:offset+Int32Key], ps.DepositTreeRoot.Bytes())
	offset += Int32Key

	copy(buf[offset:offset+Int32Key], ps.BlockHash.Bytes())

	binary.BigEndian.PutUint32(buf, ps.BlockNumber)

	return buf
}

func (ps *PublicState) Unmarshal(data []byte) error {
	if len(data) < NumPublicStateBytes {
		return errors.New("invalid data length")
	}

	offset := 0

	ps.BlockTreeRoot = new(intMaxGP.PoseidonHashOut)
	ps.BlockTreeRoot.Unmarshal(data[offset : offset+Int32Key])
	offset += Int32Key

	ps.PrevAccountTreeRoot = new(intMaxGP.PoseidonHashOut)
	ps.PrevAccountTreeRoot.Unmarshal(data[offset : offset+Int32Key])
	offset += Int32Key

	ps.AccountTreeRoot = new(intMaxGP.PoseidonHashOut)
	ps.AccountTreeRoot.Unmarshal(data[offset : offset+Int32Key])
	offset += Int32Key

	ps.DepositTreeRoot = common.BytesToHash(data[offset : offset+Int32Key])
	offset += Int32Key

	ps.BlockHash = common.BytesToHash(data[offset : offset+Int32Key])
	offset += Int32Key

	ps.BlockNumber = binary.BigEndian.Uint32(data[offset : offset+Int4Key])

	return nil
}

const (
	prevAccountTreeRootOffset = intMaxGP.NUM_HASH_OUT_ELTS
	accountTreeRootOffset     = prevAccountTreeRootOffset + intMaxGP.NUM_HASH_OUT_ELTS
	depositTreeRootOffset     = accountTreeRootOffset + intMaxGP.NUM_HASH_OUT_ELTS
	blockHashOffset           = depositTreeRootOffset + Int8Key
	blockNumberOffset         = blockHashOffset + Int8Key
	PublicStateLimbSize       = blockNumberOffset + 1
)

func (ps *PublicState) FromFieldElementSlice(value []ffg.Element) *PublicState {
	ps.BlockTreeRoot = new(intMaxGP.PoseidonHashOut).FromPartial(value[:intMaxGP.NUM_HASH_OUT_ELTS])
	ps.PrevAccountTreeRoot = new(intMaxGP.PoseidonHashOut).FromPartial(value[prevAccountTreeRootOffset:accountTreeRootOffset])
	ps.AccountTreeRoot = new(intMaxGP.PoseidonHashOut).FromPartial(value[accountTreeRootOffset:depositTreeRootOffset])
	depositTreeRoot := intMaxTypes.Bytes32{}
	copy(depositTreeRoot[:], FieldElementSliceToUint32Slice(value[depositTreeRootOffset:blockHashOffset]))
	ps.DepositTreeRoot = common.Hash{}
	copy(ps.DepositTreeRoot[:], depositTreeRoot.Bytes())
	blockHash := intMaxTypes.Bytes32{}
	copy(blockHash[:], FieldElementSliceToUint32Slice(value[blockHashOffset:blockNumberOffset]))
	ps.BlockHash = common.Hash{}
	copy(ps.BlockHash[:], blockHash.Bytes())
	ps.BlockNumber = uint32(value[blockNumberOffset].ToUint64Regular())

	return ps
}
