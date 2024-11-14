package intmax_block_content

import (
	"encoding/binary"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type PostedBlock struct {
	// The previous block hash.
	PrevBlockHash common.Hash `json:"prevBlockHash"`
	// The block number, which is the latest block number in the Rollup contract plus 1.
	BlockNumber uint32 `json:"blockNumber"`
	// The deposit root at the time of block posting (written in the Rollup contract).
	DepositRoot common.Hash `json:"depositTreeRoot"`
	// The counter of deposit leaves by current block number
	DepositLeavesCounter uint32 `json:"depositLeavesCounter"`
	// The hash value that the Block Builder must provide to the Rollup contract when posting a new block.
	SignatureHash common.Hash `json:"signatureHash"`
}

func NewPostedBlock(prevBlockHash, depositRoot common.Hash, blockNumber uint32, signatureHash common.Hash) *PostedBlock {
	return &PostedBlock{
		PrevBlockHash: prevBlockHash,
		BlockNumber:   blockNumber,
		DepositRoot:   depositRoot,
		SignatureHash: signatureHash,
	}
}

func (pb *PostedBlock) Set(other *PostedBlock) *PostedBlock {
	pb.PrevBlockHash = other.PrevBlockHash
	copy(pb.PrevBlockHash[:], other.PrevBlockHash[:])
	pb.DepositRoot = other.DepositRoot
	copy(pb.DepositRoot[:], other.DepositRoot[:])
	pb.SignatureHash = other.SignatureHash
	copy(pb.SignatureHash[:], other.SignatureHash[:])
	pb.BlockNumber = other.BlockNumber

	return pb
}

func (pb *PostedBlock) Equals(other *PostedBlock) bool {
	return pb.PrevBlockHash == other.PrevBlockHash &&
		pb.DepositRoot == other.DepositRoot &&
		pb.SignatureHash == other.SignatureHash &&
		pb.BlockNumber == other.BlockNumber
}

func (pb *PostedBlock) Genesis() *PostedBlock {
	depositTree, err := intMaxTree.NewDepositTree(intMaxTree.DEPOSIT_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}

	depositTreeRoot, _, _ := depositTree.GetCurrentRootCountAndSiblings()

	return NewPostedBlock(common.Hash{}, depositTreeRoot, 0, common.Hash{})
}

func (pb *PostedBlock) Marshal() []byte {
	const int4Key = 4

	data := make([]byte, 0)

	data = append(data, pb.PrevBlockHash.Bytes()...)
	data = append(data, pb.DepositRoot.Bytes()...)
	data = append(data, pb.SignatureHash.Bytes()...)
	blockNumberBytes := [int4Key]byte{}
	binary.BigEndian.PutUint32(blockNumberBytes[:], pb.BlockNumber)
	data = append(data, blockNumberBytes[:]...)

	return data
}

func (pb *PostedBlock) Uint32Slice() []uint32 {
	var buf []uint32
	buf = append(buf, intMaxTypes.CommonHashToUint32Slice(pb.PrevBlockHash)...)
	buf = append(buf, intMaxTypes.CommonHashToUint32Slice(pb.DepositRoot)...)
	buf = append(buf, intMaxTypes.CommonHashToUint32Slice(pb.SignatureHash)...)
	buf = append(buf, pb.BlockNumber)

	return buf
}

func (pb *PostedBlock) Hash() common.Hash {
	return crypto.Keccak256Hash(intMaxTypes.Uint32SliceToBytes(pb.Uint32Slice()))
}
