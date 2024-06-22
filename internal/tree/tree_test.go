package tree

import (
	"intmax2-node/internal/hash/goldenposeidon"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPoseidonMerkleTreeWithoutInitialLeaves(t *testing.T) {
	zero := goldenposeidon.NewPoseidonHashOut()
	initialLeaves := make([]*poseidonHashOut, 0)
	mt, err := NewPoseidonMerkleTree(3, initialLeaves, zero)
	if err != nil {
		t.Errorf("fail to create merkle tree")
	}

	leaves := make([]*poseidonHashOut, 8)
	for i := 0; i < 4; i++ {
		leaves[i], _ = new(poseidonHashOut).SetRandom()
		_, err := mt.AddLeaf(uint64(i), leaves[i])
		if err != nil {
			t.Errorf("fail to add leaf")
		}
	}

	expectedRoot := goldenposeidon.Compress(
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[0], leaves[1]), goldenposeidon.Compress(leaves[2], leaves[3])),
		goldenposeidon.Compress(goldenposeidon.Compress(zero, zero), goldenposeidon.Compress(zero, zero)),
	)
	assert.Equal(t, expectedRoot.Elements, mt.currentRoot.Elements)

	leaves[4], _ = new(poseidonHashOut).SetRandom()
	_, err = mt.AddLeaf(4, leaves[4])
	if err != nil {
		t.Errorf("fail to add leaf")
	}

	expectedRoot = goldenposeidon.Compress(
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[0], leaves[1]), goldenposeidon.Compress(leaves[2], leaves[3])),
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[4], zero), goldenposeidon.Compress(zero, zero)),
	)
	assert.Equal(t, expectedRoot.Elements, mt.currentRoot.Elements)

	for i := 5; i < 8; i++ {
		leaves[i], _ = new(poseidonHashOut).SetRandom()
		_, err := mt.AddLeaf(uint64(i), leaves[i])
		if err != nil {
			t.Errorf("fail to add leaf")
		}
	}

	expectedRoot = goldenposeidon.Compress(
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[0], leaves[1]), goldenposeidon.Compress(leaves[2], leaves[3])),
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[4], leaves[5]), goldenposeidon.Compress(leaves[6], leaves[7])),
	)
	assert.Equal(t, expectedRoot.Elements, mt.currentRoot.Elements)
}

func TestPoseidonMerkleTreeWithInitialLeaves(t *testing.T) {
	leaves := make([]*poseidonHashOut, 8)

	zero := goldenposeidon.NewPoseidonHashOut()
	initialLeaves := make([]*poseidonHashOut, 3)
	for i := 0; i < len(initialLeaves); i++ {
		leaves[i], _ = new(poseidonHashOut).SetRandom()
		initialLeaves[i] = leaves[i]
	}
	mt, err := NewPoseidonMerkleTree(3, initialLeaves, zero)
	if err != nil {
		t.Errorf("fail to create merkle tree")
	}

	expectedRoot := goldenposeidon.Compress(
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[0], leaves[1]), goldenposeidon.Compress(leaves[2], zero)),
		goldenposeidon.Compress(goldenposeidon.Compress(zero, zero), goldenposeidon.Compress(zero, zero)),
	)
	assert.Equal(t, expectedRoot.Elements, mt.currentRoot.Elements)

	leaves[3], _ = new(poseidonHashOut).SetRandom()
	_, err = mt.AddLeaf(3, leaves[3])
	if err != nil {
		t.Errorf("fail to add leaf")
	}

	expectedRoot = goldenposeidon.Compress(
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[0], leaves[1]), goldenposeidon.Compress(leaves[2], leaves[3])),
		goldenposeidon.Compress(goldenposeidon.Compress(zero, zero), goldenposeidon.Compress(zero, zero)),
	)
	assert.Equal(t, expectedRoot.Elements, mt.currentRoot.Elements)

	leaves[4], _ = new(poseidonHashOut).SetRandom()
	_, err = mt.AddLeaf(4, leaves[4])
	if err != nil {
		t.Errorf("fail to add leaf")
	}

	expectedRoot = goldenposeidon.Compress(
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[0], leaves[1]), goldenposeidon.Compress(leaves[2], leaves[3])),
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[4], zero), goldenposeidon.Compress(zero, zero)),
	)
	assert.Equal(t, expectedRoot.Elements, mt.currentRoot.Elements)

	for i := 5; i < 8; i++ {
		leaves[i], _ = new(poseidonHashOut).SetRandom()
		_, err := mt.AddLeaf(uint64(i), leaves[i])
		if err != nil {
			t.Errorf("fail to add leaf")
		}
	}

	expectedRoot = goldenposeidon.Compress(
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[0], leaves[1]), goldenposeidon.Compress(leaves[2], leaves[3])),
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[4], leaves[5]), goldenposeidon.Compress(leaves[6], leaves[7])),
	)
	assert.Equal(t, expectedRoot.Elements, mt.currentRoot.Elements)
}

func TestComputeMerkleProof(t *testing.T) {
	zero := goldenposeidon.NewPoseidonHashOut()
	initialLeaves := make([]*poseidonHashOut, 0)
	mt, err := NewPoseidonMerkleTree(5, initialLeaves, zero)
	require.NoError(t, err)
	leaves := []*poseidonHashOut{
		goldenposeidon.HexToHash("0x83fc198de31e1b2b1a8212d2430fbb7766c13d9ad305637dea3759065606475d"),
		goldenposeidon.HexToHash("0x0349657c7850dc9b2b73010501b01cd6a38911b6a2ad2167c164c5b2a5b344de"),
		goldenposeidon.HexToHash("0xb32f96fad8af99f3b3cb90dfbb4849f73435dbee1877e4ac2c213127379549ce"),
		goldenposeidon.HexToHash("0x79ffa1294bf48e0dd41afcb23b2929921e4e17f2f81b7163c23078375b06ba4f"),
		goldenposeidon.HexToHash("0x0004063b5c83f56a17f580db0908339c01206cdf8b59beb13ce6f146bb025fe2"),
		goldenposeidon.HexToHash("0x68e4f2c517c7f60c3664ac6bbe78f904eacdbe84790aa0d15d79ddd6216c556e"),
		goldenposeidon.HexToHash("0xf7245f4d84367a189b90873e4563a000702dbfe974b872fdb13323a828c8fb71"),
	}
	siblings, root, err := mt.ComputeMerkleProof(1, leaves)
	require.NoError(t, err)
	require.Equal(t, "0xad1d6575a142d2bb0dfe324d17ccc36df4719e5adbc546c11310524d8aa922cb", root.String())
	expectedProof := []string{"0x83fc198de31e1b2b1a8212d2430fbb7766c13d9ad305637dea3759065606475d", "0x541aa343e422c6af36eccf5fbc687aa58d9de38364398e106715e816904ea7f8", "0x84108b427bf14dcdcd31f8090a6f8d60b88e00bb5f110474abf801b6e0a568ec", "0x5ae05c29f70ae06164dea29dc57c249a5fc056e9bf94fb4642a53cc70c3a7067", "0x442646061a92545147092c2e0db3c18c274d85bff37c7d1640a088afa0ea22f5"}
	for i := 0; i < len(siblings); i++ {
		require.Equal(t, expectedProof[i], siblings[i].String())
	}
}
