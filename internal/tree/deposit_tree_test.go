package tree

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// hash: 79bdfe22442dfca3237f7d073c36739ae9c03d76b8cb995f61ec826311efae65
// depositLeaf.Hash(): 0x79bdfe22442dfca3237f7d073c36739ae9c03d76b8cb995f61ec826311efae65
// root: 0x484bef7ed59d1224ccdcbf8d178096b894e6f0d6217b27390d7f928139725ac7, count: 3
// packed: d5e60fdf88a83d7f4729fc7e7ffe55fe97a94e697231c7e2242f69ba2077a216000000030000000000000000000000000000000000000000000000000000000000002710
// hash: 79bdfe22442dfca3237f7d073c36739ae9c03d76b8cb995f61ec826311efae65
// lastDepositRoot: 0x77a767863b78c81f2323f84811b2018f34e4ee85ddd225b73e58b663764b65b2
// root: 0x77a767863b78c81f2323f84811b2018f34e4ee85ddd225b73e58b663764b65b2, count: 4
// Processing deposits from block 5927829, depositId 6
// lastDepositRoot: 0x77a767863b78c81f2323f84811b2018f34e4ee85ddd225b73e58b663764b65b2
// depositLeaf.RecipientSaltHash: 786fb5cda92cfa2f266ba4f4d13b3861211686771a92e6d31f3356eee58cfe82
// depositLeaf.TokenIndex: 3
// depositLeaf.Amount: 1
// packed: 786fb5cda92cfa2f266ba4f4d13b3861211686771a92e6d31f3356eee58cfe82000000030000000000000000000000000000000000000000000000000000000000000001
// hash: 45985d70ecf86d517f94157973e0359ec6461aa561628e526bb77f34b175696f
// depositLeaf.Hash(): 0x45985d70ecf86d517f94157973e0359ec6461aa561628e526bb77f34b175696f
// root: 0x77a767863b78c81f2323f84811b2018f34e4ee85ddd225b73e58b663764b65b2, count: 4
// packed: 786fb5cda92cfa2f266ba4f4d13b3861211686771a92e6d31f3356eee58cfe82000000030000000000000000000000000000000000000000000000000000000000000001
// hash: 45985d70ecf86d517f94157973e0359ec6461aa561628e526bb77f34b175696f
// lastDepositRoot: 0x87dca48d9badaf63339bc4be9cbf60aad97838d121514701d7d4b98e723b263c
// root: 0x87dca48d9badaf63339bc4be9cbf60aad97838d121514701d7d4b98e723b263c, count: 5
// depositLeaf.RecipientSaltHash: f645283a99835525833b0b2f56373a46001dff7609b8b5dc2443ffeb492d0fe6
// depositLeaf.TokenIndex: 3
// depositLeaf.Amount: 10000
// packed: f645283a99835525833b0b2f56373a46001dff7609b8b5dc2443ffeb492d0fe6000000030000000000000000000000000000000000000000000000000000000000002710
// hash: 75872ec90258b5998bbed4bb989de7794b2c1e99beecc9e7ebc8bff7a96a26b0
// depositLeaf.Hash(): 0x75872ec90258b5998bbed4bb989de7794b2c1e99beecc9e7ebc8bff7a96a26b0
// root: 0x87dca48d9badaf63339bc4be9cbf60aad97838d121514701d7d4b98e723b263c, count: 5
// packed: f645283a99835525833b0b2f56373a46001dff7609b8b5dc2443ffeb492d0fe6000000030000000000000000000000000000000000000000000000000000000000002710
// hash: 75872ec90258b5998bbed4bb989de7794b2c1e99beecc9e7ebc8bff7a96a26b0
// lastDepositRoot: 0xa421b6b0e2872549fe7d221b0dbd793820152b1058bb6bd71f323b368e44804e
// root: 0xa421b6b0e2872549fe7d221b0dbd793820152b1058bb6bd71f323b368e44804e, count: 6
// Syncing deposits from block 5943354
// Found 9 deposits processed events
// Processing deposits from block 5943794, depositId 7
// lastDepositRoot: 0xa421b6b0e2872549fe7d221b0dbd793820152b1058bb6bd71f323b368e44804e
// depositLeaf.RecipientSaltHash: 02cdae12f0c152a32a3a620060397614fd671388296a2d024a9d6b2ef471f213
// depositLeaf.TokenIndex: 0
// depositLeaf.Amount: 1000000000000000
// packed: 02cdae12f0c152a32a3a620060397614fd671388296a2d024a9d6b2ef471f2130000000000000000000000000000000000000000000000000000000000038d7ea4c68000
// hash: aa9ef12496331dc2acc26c6420f081caf803a748e5908a2d9f5f1e0ea4399d49
// depositLeaf.Hash(): 0xaa9ef12496331dc2acc26c6420f081caf803a748e5908a2d9f5f1e0ea4399d49
// root: 0xa421b6b0e2872549fe7d221b0dbd793820152b1058bb6bd71f323b368e44804e, count: 6
// packed: 02cdae12f0c152a32a3a620060397614fd671388296a2d024a9d6b2ef471f2130000000000000000000000000000000000000000000000000000000000038d7ea4c68000
// hash: aa9ef12496331dc2acc26c6420f081caf803a748e5908a2d9f5f1e0ea4399d49
// lastDepositRoot: 0x7c24c2a267c9415f6fc0fe161e0903de154a21ca2eedf54c9747d9d15d78342b
// root: 0x7c24c2a267c9415f6fc0fe161e0903de154a21ca2eedf54c9747d9d15d78342b, count: 7
// depositLeaf.RecipientSaltHash: cde67779b90f1b18d215fc9023553ca2a66c677d7bc403eb3cb4f6c3cbb511c6
// depositLeaf.TokenIndex: 0
// depositLeaf.Amount: 10
// packed: cde67779b90f1b18d215fc9023553ca2a66c677d7bc403eb3cb4f6c3cbb511c600000000000000000000000000000000000000000000000000000000000000000000000a
// hash: 9885604ad4795a015af9233a292a0568b9b4a921d72501698aa8958aeb256253
// depositLeaf.Hash(): 0x9885604ad4795a015af9233a292a0568b9b4a921d72501698aa8958aeb256253
// root: 0x7c24c2a267c9415f6fc0fe161e0903de154a21ca2eedf54c9747d9d15d78342b, count: 7
// packed: cde67779b90f1b18d215fc9023553ca2a66c677d7bc403eb3cb4f6c3cbb511c600000000000000000000000000000000000000000000000000000000000000000000000a
// hash: 9885604ad4795a015af9233a292a0568b9b4a921d72501698aa8958aeb256253
// lastDepositRoot: 0x79c078ca4902b859522de1faf67ccdad9b440e661a39e86da2fa1ddb1b843b3d
// root: 0x79c078ca4902b859522de1faf67ccdad9b440e661a39e86da2fa1ddb1b843b3d, count: 8
// depositLeaf.RecipientSaltHash: 3ec96cade886ad7c82c58c5cc3e0e3613814b2257156d70f3ab622c0a28a8fb0
// depositLeaf.TokenIndex: 0
// depositLeaf.Amount: 10000
// packed: 3ec96cade886ad7c82c58c5cc3e0e3613814b2257156d70f3ab622c0a28a8fb0000000000000000000000000000000000000000000000000000000000000000000002710
// hash: 67a8ebf9c746d1e1badfc8e65621715f32654f061f3314fa22a778c2b70686cc
// depositLeaf.Hash(): 0x67a8ebf9c746d1e1badfc8e65621715f32654f061f3314fa22a778c2b70686cc
// root: 0x79c078ca4902b859522de1faf67ccdad9b440e661a39e86da2fa1ddb1b843b3d, count: 8
// packed: 3ec96cade886ad7c82c58c5cc3e0e3613814b2257156d70f3ab622c0a28a8fb0000000000000000000000000000000000000000000000000000000000000000000002710
// hash: 67a8ebf9c746d1e1badfc8e65621715f32654f061f3314fa22a778c2b70686cc
// lastDepositRoot: 0x3449bbdd6cff8683e9633fd0e65b3d3630fa70ba054de5e2aefaa496bd63a1c8
// root: 0x3449bbdd6cff8683e9633fd0e65b3d3630fa70ba054de5e2aefaa496bd63a1c8, count: 9
// depositLeaf.RecipientSaltHash: d5e60fdf88a83d7f4729fc7e7ffe55fe97a94e697231c7e2242f69ba2077a216
// depositLeaf.TokenIndex: 3
// depositLeaf.Amount: 10000
// packed: d5e60fdf88a83d7f4729fc7e7ffe55fe97a94e697231c7e2242f69ba2077a216000000030000000000000000000000000000000000000000000000000000000000002710
// hash: 79bdfe22442dfca3237f7d073c36739ae9c03d76b8cb995f61ec826311efae65
// depositLeaf.Hash(): 0x79bdfe22442dfca3237f7d073c36739ae9c03d76b8cb995f61ec826311efae65
// root: 0x3449bbdd6cff8683e9633fd0e65b3d3630fa70ba054de5e2aefaa496bd63a1c8, count: 9
// packed: d5e60fdf88a83d7f4729fc7e7ffe55fe97a94e697231c7e2242f69ba2077a216000000030000000000000000000000000000000000000000000000000000000000002710
// hash: 79bdfe22442dfca3237f7d073c36739ae9c03d76b8cb995f61ec826311efae65
// lastDepositRoot: 0xdc57276e70e98f980dee5bf4f392c1004e190e8e22e06a9a7585f384c7c8dd20
// root: 0xdc57276e70e98f980dee5bf4f392c1004e190e8e22e06a9a7585f384c7c8dd20, count: 10
// depositLeaf.RecipientSaltHash: 786fb5cda92cfa2f266ba4f4d13b3861211686771a92e6d31f3356eee58cfe82
// depositLeaf.TokenIndex: 3
// depositLeaf.Amount: 1
// packed: 786fb5cda92cfa2f266ba4f4d13b3861211686771a92e6d31f3356eee58cfe82000000030000000000000000000000000000000000000000000000000000000000000001
// hash: 45985d70ecf86d517f94157973e0359ec6461aa561628e526bb77f34b175696f
// depositLeaf.Hash(): 0x45985d70ecf86d517f94157973e0359ec6461aa561628e526bb77f34b175696f
// root: 0xdc57276e70e98f980dee5bf4f392c1004e190e8e22e06a9a7585f384c7c8dd20, count: 10
// packed: 786fb5cda92cfa2f266ba4f4d13b3861211686771a92e6d31f3356eee58cfe82000000030000000000000000000000000000000000000000000000000000000000000001
// hash: 45985d70ecf86d517f94157973e0359ec6461aa561628e526bb77f34b175696f
// lastDepositRoot: 0x7d197e6baa65075f7baeab4207795f36d0c2fb2f461553e9dc7050e75ebd870e
// root: 0x7d197e6baa65075f7baeab4207795f36d0c2fb2f461553e9dc7050e75ebd870e, count: 11
// depositLeaf.RecipientSaltHash: f645283a99835525833b0b2f56373a46001dff7609b8b5dc2443ffeb492d0fe6
// depositLeaf.TokenIndex: 3
// depositLeaf.Amount: 10000
// packed: f645283a99835525833b0b2f56373a46001dff7609b8b5dc2443ffeb492d0fe6000000030000000000000000000000000000000000000000000000000000000000002710
// hash: 75872ec90258b5998bbed4bb989de7794b2c1e99beecc9e7ebc8bff7a96a26b0
// depositLeaf.Hash(): 0x75872ec90258b5998bbed4bb989de7794b2c1e99beecc9e7ebc8bff7a96a26b0
// root: 0x7d197e6baa65075f7baeab4207795f36d0c2fb2f461553e9dc7050e75ebd870e, count: 11
// packed: f645283a99835525833b0b2f56373a46001dff7609b8b5dc2443ffeb492d0fe6000000030000000000000000000000000000000000000000000000000000000000002710
// hash: 75872ec90258b5998bbed4bb989de7794b2c1e99beecc9e7ebc8bff7a96a26b0
// lastDepositRoot: 0x982eb79f49e0bc6a57461aad7d6887c94b957892c3480335a5cf38a6112f733a
// root: 0x982eb79f49e0bc6a57461aad7d6887c94b957892c3480335a5cf38a6112f733a, count: 12
// depositLeaf.RecipientSaltHash: e0b5e1d78455a700efa098be8f35594ecbc25b020124680a06ce8e7ba4e8a3b1
// depositLeaf.TokenIndex: 3
// depositLeaf.Amount: 10000
// packed: e0b5e1d78455a700efa098be8f35594ecbc25b020124680a06ce8e7ba4e8a3b1000000030000000000000000000000000000000000000000000000000000000000002710
// hash: e25a53597bc0def623f749a499d18ee35a7e23fff706695981e1d8c451572e18
// depositLeaf.Hash(): 0xe25a53597bc0def623f749a499d18ee35a7e23fff706695981e1d8c451572e18
// root: 0x982eb79f49e0bc6a57461aad7d6887c94b957892c3480335a5cf38a6112f733a, count: 12
// packed: e0b5e1d78455a700efa098be8f35594ecbc25b020124680a06ce8e7ba4e8a3b1000000030000000000000000000000000000000000000000000000000000000000002710
// hash: e25a53597bc0def623f749a499d18ee35a7e23fff706695981e1d8c451572e18
// lastDepositRoot: 0x6db2847aa710f29f822c203ec9f7873d51bfd08bc60e5182b52604091c21d249
// root: 0x6db2847aa710f29f822c203ec9f7873d51bfd08bc60e5182b52604091c21d249, count: 13

func TestDepositTreeWithoutInitialLeaves(t *testing.T) {
	mt, err := NewDepositTree(32)
	if err != nil {
		t.Errorf("fail to create merkle tree")
	}

	var recipientSaltHash []byte

	type RawDepositLeaf struct {
		DepositId         uint32
		RecipientSaltHash string
		TokenIndex        uint32
		Amount            string
	}

	rawLeaves := []RawDepositLeaf{
		{
			DepositId:         0,
			RecipientSaltHash: "02cdae12f0c152a32a3a620060397614fd671388296a2d024a9d6b2ef471f213",
			TokenIndex:        0,
			Amount:            "1000000000000000",
		},
		{
			DepositId:         1,
			RecipientSaltHash: "cde67779b90f1b18d215fc9023553ca2a66c677d7bc403eb3cb4f6c3cbb511c6",
			TokenIndex:        0,
			Amount:            "10",
		},
		{
			DepositId:         2,
			RecipientSaltHash: "3ec96cade886ad7c82c58c5cc3e0e3613814b2257156d70f3ab622c0a28a8fb0",
			TokenIndex:        0,
			Amount:            "10000",
		},
		{
			DepositId:         3,
			RecipientSaltHash: "d5e60fdf88a83d7f4729fc7e7ffe55fe97a94e697231c7e2242f69ba2077a216",
			TokenIndex:        3,
			Amount:            "10000",
		},
		{
			DepositId:         4,
			RecipientSaltHash: "786fb5cda92cfa2f266ba4f4d13b3861211686771a92e6d31f3356eee58cfe82",
			TokenIndex:        3,
			Amount:            "1",
		},
		{
			DepositId:         5,
			RecipientSaltHash: "f645283a99835525833b0b2f56373a46001dff7609b8b5dc2443ffeb492d0fe6",
			TokenIndex:        3,
			Amount:            "10000",
		},
		// {
		// 	RecipientSaltHash: "02cdae12f0c152a32a3a620060397614fd671388296a2d024a9d6b2ef471f213",
		// 	TokenIndex:        0,
		// 	Amount:            "1000000000000000",
		// },
		// {
		// 	RecipientSaltHash: "cde67779b90f1b18d215fc9023553ca2a66c677d7bc403eb3cb4f6c3cbb511c6",
		// 	TokenIndex:        0,
		// 	Amount:            "10",
		// },
		// {
		// 	RecipientSaltHash: "3ec96cade886ad7c82c58c5cc3e0e3613814b2257156d70f3ab622c0a28a8fb0",
		// 	TokenIndex:        0,
		// 	Amount:            "10000",
		// },
		// {
		// 	RecipientSaltHash: "d5e60fdf88a83d7f4729fc7e7ffe55fe97a94e697231c7e2242f69ba2077a216",
		// 	TokenIndex:        3,
		// 	Amount:            "10000",
		// },
		// {
		// 	RecipientSaltHash: "786fb5cda92cfa2f266ba4f4d13b3861211686771a92e6d31f3356eee58cfe82",
		// 	TokenIndex:        3,
		// 	Amount:            "1",
		// },
		// {
		// 	DepositId:         11,
		// 	RecipientSaltHash: "f645283a99835525833b0b2f56373a46001dff7609b8b5dc2443ffeb492d0fe6",
		// 	TokenIndex:        3,
		// 	Amount:            "10000",
		// },
		{
			DepositId:         12,
			RecipientSaltHash: "e0b5e1d78455a700efa098be8f35594ecbc25b020124680a06ce8e7ba4e8a3b1",
			TokenIndex:        3,
			Amount:            "10000",
		},
	}

	leaves := make([]*DepositLeaf, len(rawLeaves))
	for i, rawLeaf := range rawLeaves {
		leaf := DepositLeaf{
			RecipientSaltHash: [32]byte{},
			TokenIndex:        rawLeaf.TokenIndex,
			Amount:            new(big.Int),
		}
		recipientSaltHash, err = hex.DecodeString(rawLeaf.RecipientSaltHash)
		require.NoError(t, err)
		copy(leaf.RecipientSaltHash[:], recipientSaltHash)
		_, ok := leaf.Amount.SetString(rawLeaf.Amount, 10)
		require.True(t, ok)

		leaves[i] = &leaf
	}

	for i, leaf := range leaves {
		fmt.Printf("--- leaves[%d] ---\n", i)
		fmt.Printf("deposit ID: %d\n", rawLeaves[i].DepositId)

		_, count, siblings := mt.GetCurrentRootCountAndSiblings()
		fmt.Printf("Siblings: %x\n", siblings)

		expectedRoot, err := mt.AddLeaf(count, *leaf)
		require.NoError(t, err)

		require.Equal(t, common.Hash(expectedRoot), mt.inner.currentRoot)
		fmt.Printf("Root: %x\n", mt.inner.currentRoot)
	}

	// root after inserted leaves[6]: 7c24c2a267c9415f6fc0fe161e0903de154a21ca2eedf54c9747d9d15d78342b, but got 19e091301c6758d623167ce8876f1477acbb44350f758a7e8e564b4464930bbf
}
