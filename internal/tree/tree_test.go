package tree_test

import (
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	tree "intmax2-node/internal/tree"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMerkleProof(t *testing.T) {
	leaves := make([]intMaxGP.PoseidonHashOut, 2)
	err := leaves[0].FromString("0x545cac70c52cf8589c16de1eb85e264d51e18adb15ac810db3f44efa190a1074")
	require.Nil(t, err)
	err = leaves[1].FromString("0x4b44d51735ffd85fa54d6c3cc60352648ab093840fe4095b39afee145bf0c367")
	require.Nil(t, err)
	defaultLeafHash := new(intMaxGP.PoseidonHashOut)

	height := uint8(2)
	mt, err := tree.NewPoseidonMerkleTree(
		height,
		defaultLeafHash,
	)
	require.Nil(t, err)

	for i, leaf := range leaves {
		mt.UpdateLeaf(i, &leaf)
	}

	root := mt.GetRoot()
	t.Logf("tx tree root: %x\n", root.Marshal())

	for i := 1; i < 1<<(height+1); i++ {
		h := mt.GetNodeHash(i)
		t.Logf("nodeHashes[%d]: %x\n", i, h.Marshal())
	}

	index := 0
	proof, err := mt.Prove(index)
	require.Nil(t, err)
	leaf := mt.GetLeaf(index)

	for i, sibling := range proof.Siblings {
		t.Logf("proof[%d]: %x\n", i, sibling.Marshal())
	}

	err = proof.Verify(leaf, index, root)
	require.Nil(t, err)
}
