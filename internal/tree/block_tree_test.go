package tree_test

import (
	"errors"
	"fmt"
	"intmax2-node/internal/block_post_service"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	tree "intmax2-node/internal/tree"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

const BLOCK_HASH_TREE_HEIGHT = 32

type PoseidonHashOut = intMaxGP.PoseidonHashOut

type BlockHashMerkleProof = tree.PoseidonMerkleProof

type BlockHashTree struct {
	Leaves          []*tree.BlockHashLeaf
	NextBlockNumber uint32
	inner           *tree.PoseidonMerkleTree
}

func (t *BlockHashTree) Set(other *BlockHashTree) *BlockHashTree {
	t.Leaves = make([]*tree.BlockHashLeaf, len(other.Leaves))
	copy(t.Leaves, other.Leaves)
	t.inner = new(tree.PoseidonMerkleTree).Set(other.inner)

	return t
}

func NewBlockHashTreeWithInitialLeaves(height uint8, initialLeaves []*tree.BlockHashLeaf) (*BlockHashTree, error) {
	zeroHash := new(tree.BlockHashLeaf).SetDefault().Hash()
	t, err := tree.NewPoseidonMerkleTree(height, zeroHash)
	if err != nil {
		return nil, err
	}

	for i, leaf := range initialLeaves {
		t.UpdateLeaf(i, leaf.Hash())
	}

	leaves := make([]*tree.BlockHashLeaf, len(initialLeaves))
	copy(leaves, initialLeaves)

	return &BlockHashTree{
		Leaves:          leaves,
		NextBlockNumber: uint32(len(initialLeaves)),
		inner:           t,
	}, nil
}

// func (t *BlockHashTree) BuildMerkleRoot(leaves [][numHashBytes]byte) (*PoseidonHashOut, error) {
// 	initialLeaves := make([]*PoseidonHashOut, len(leaves))
// 	for i, leaf := range leaves {
// 		initialLeaves[i] = NewBlockHashLeaf(leaf).Hash()
// 	}

// 	return t.inner.BuildMerkleRoot(initialLeaves)
// }

func (t *BlockHashTree) GetCurrentRootCountAndSiblings() (root tree.PoseidonHashOut, nextIndex uint32, siblings []*tree.PoseidonHashOut) {
	// root, count, siblings := t.inner.GetCurrentRootCountAndSiblings()
	root = *t.inner.GetRoot()
	nextIndex = t.NextBlockNumber
	proof, err := t.inner.Prove(int(nextIndex))
	if err != nil {
		panic(err)
	}

	return root, nextIndex, proof.Siblings
}

func (t *BlockHashTree) AddLeaf(index uint32, leaf *tree.BlockHashLeaf) (root *PoseidonHashOut, err error) {
	leafHash := leaf.Hash()
	t.inner.UpdateLeaf(int(index), leafHash)
	// if err != nil {
	// 	return nil, err
	// }

	if int(index) != len(t.Leaves) {
		return nil, errors.New("index is not equal to the length of block leaves")
	}

	t.Leaves = append(t.Leaves, leaf)
	t.NextBlockNumber++

	return root, nil
}

func (t *BlockHashTree) ComputeMerkleProof(index uint32, leaves []tree.BlockHashLeaf) (siblings []*PoseidonHashOut, root PoseidonHashOut, err error) {
	leafHashes := make([]*PoseidonHashOut, len(leaves))
	for i, leaf := range leaves {
		leafHashes[i] = leaf.Hash()
	}

	proof, err := t.inner.Prove(int(index))
	return proof.Siblings, *t.inner.GetRoot(), err
}

func (t *BlockHashTree) Prove(index uint32) (proof tree.PoseidonMerkleProof, root *PoseidonHashOut, leafHash *PoseidonHashOut, err error) {
	proof, err = t.inner.Prove(int(index))
	if err != nil {
		return proof, root, leafHash, err
	}

	root = t.inner.GetRoot()
	leafHash = t.inner.GetLeaf(int(index))

	err = proof.Verify(leafHash, int(index), root)
	if err != nil {
		panic(fmt.Errorf("Fatal Error: Merkle proof verification failed: %w", err))
	}

	if !leafHash.Equal(t.GetLeaf(index).Hash()) {
		panic(fmt.Errorf("Fatal Error: leafHash is mismatch %s != %s", leafHash.String(), t.GetLeaf(index).Hash().String()))
	}

	return proof, root, leafHash, err
}

func (t *BlockHashTree) GetRoot() *PoseidonHashOut {
	return t.inner.GetRoot()
}

func (t *BlockHashTree) GetLeaf(index uint32) *tree.BlockHashLeaf {
	if int(index) >= len(t.Leaves) {
		return new(tree.BlockHashLeaf).SetDefault()
	}

	return t.Leaves[index]
}

func TestBlockTreeProof(t *testing.T) {
	genesisBlock := new(block_post_service.PostedBlock).Genesis()
	blockLeaf0 := tree.NewBlockHashLeaf(genesisBlock.Hash())
	blockHash1 := common.HexToHash("0x4b44d51735ffd85fa54d6c3cc60352648ab093840fe4095b39afee145bf0c367")
	blockLeaf1 := tree.NewBlockHashLeaf(blockHash1)

	blockTree, err := NewBlockHashTreeWithInitialLeaves(2, []*tree.BlockHashLeaf{blockLeaf0})
	require.NoError(t, err)
	proof, root, leafHash, err := blockTree.Prove(0)
	require.NoError(t, err)
	t.Log("blockTree root:", blockTree.GetRoot().String())
	t.Log("blockTree root:", root.String())
	t.Log("blockTree proof:", proof)
	// leaf0 := blockTree.GetLeaf(0)
	err = proof.Verify(leafHash, 0, root)
	require.NoError(t, err)
	err = proof.Verify(blockLeaf0.Hash(), 0, root)
	require.NoError(t, err)

	proof, root, _, err = blockTree.Prove(1)
	require.NoError(t, err)
	leaf1 := blockTree.GetLeaf(1)
	err = proof.Verify(leaf1.Hash(), 1, root)
	require.NoError(t, err)
	_, err = blockTree.AddLeaf(1, blockLeaf1)
	require.NoError(t, err)
	t.Log("blockTree root:", blockTree.GetRoot())
	err = proof.Verify(blockLeaf1.Hash(), 1, blockTree.GetRoot())
	require.NoError(t, err)
}
