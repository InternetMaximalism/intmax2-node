package block_builder_storage

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	bbsTypes "intmax2-node/internal/block_builder_storage/types"
	"intmax2-node/internal/block_post_service"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type blockBuilderStorage struct {
	cfg                      *configs.Config
	log                      logger.Logger
	latestWitnessBlockNumber uint32
	AccountTree              *intMaxTree.AccountTree
	BlockTree                *intMaxTree.BlockHashTree
	DepositTree              *intMaxTree.KeccakMerkleTree
	DepositLeaves            []*intMaxTree.DepositLeaf
	MerkleTreeHistory        map[uint32]*bbsTypes.MerkleTrees
}

func NewBlockBuilderStorage(cfg *configs.Config, log logger.Logger) (BlockBuilderStorage, error) {
	bbs := blockBuilderStorage{
		cfg: cfg,
		log: log,
	}

	return &bbs, nil
}

func (bbs *blockBuilderStorage) Init(db SQLDriverApp) (err error) {
	const (
		int0Key = 0
		int1Key = 1
		int2Key = 2
	)

	var lastProofGeneratedBlockNumber uint32
	lastProofGeneratedBlockNumber, err = db.LastBlockNumberGeneratedValidityProof()
	if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
		return errors.Join(ErrLastBlockNumberGeneratedValidityProofFail, err)
	}

	merkleTreeHistory := make(map[uint32]*bbsTypes.MerkleTrees)

	var (
		blockHashAndSendersMap      map[uint32]mDBApp.BlockHashAndSenders
		lastSynchronizedBlockNumber uint32
	)
	blockHashAndSendersMap, lastSynchronizedBlockNumber, err = db.ScanBlockHashAndSenders()
	if err != nil {
		return errors.Join(ErrScanBlockHashAndSendersFailed, err)
	}

	var accountTree *intMaxTree.AccountTree
	accountTree, err = intMaxTree.NewAccountTree(intMaxTree.ACCOUNT_TREE_HEIGHT)
	if err != nil {
		return errors.Join(ErrNewAccountTreeFail, err)
	}

	genesisBlock := new(block_post_service.PostedBlock).Genesis()
	var blockHashTree *intMaxTree.BlockHashTree
	blockHashTree, err = intMaxTree.NewBlockHashTreeWithInitialLeaves(
		intMaxTree.BLOCK_HASH_TREE_HEIGHT,
		[]*intMaxTree.BlockHashLeaf{
			intMaxTree.NewBlockHashLeaf(genesisBlock.Hash()),
		},
	)
	if err != nil {
		return errors.Join(ErrNewBlockHashTreeWithInitialLeavesFail, err)
	}

	zeroDepositHash := new(intMaxTree.DepositLeaf).SetZero().Hash()

	var depositTree *intMaxTree.KeccakMerkleTree
	depositTree, err = intMaxTree.NewKeccakMerkleTree(
		intMaxTree.DEPOSIT_TREE_HEIGHT, nil, zeroDepositHash,
	)
	if err != nil {
		return errors.Join(ErrNewKeccakMerkleTreeFail, err)
	}

	merkleTreeHistory[0] = &bbsTypes.MerkleTrees{
		AccountTree:   new(intMaxTree.AccountTree).Set(accountTree),
		BlockHashTree: new(intMaxTree.BlockHashTree).Set(blockHashTree),
		DepositLeaves: make([]*intMaxTree.DepositLeaf, int0Key),
	}

	bbs.log.Debugf("blockHashTree 0 root: %s", blockHashTree.GetRoot().String())

	blockHashes := make([]*intMaxTree.BlockHashLeaf, lastSynchronizedBlockNumber+int1Key)
	defaultPublicKey := new(intMaxAcc.Address).String()                  // zero
	dummyPublicKey := intMaxAcc.NewDummyPublicKey().ToAddress().String() // one
	for blockNumber := uint32(int1Key); blockNumber <= lastSynchronizedBlockNumber; blockNumber++ {
		blockHashAndSenders, ok := blockHashAndSendersMap[blockNumber]
		if !ok {
			const msg = "block number %d not found"
			err = fmt.Errorf(msg, blockNumber)
			return errors.Join(ErrBlockNumberNotFoundWithBlockHashAndSendersMap, err)
		}

		merkleTreeHistory[blockNumber] = new(bbsTypes.MerkleTrees)

		bbs.log.Debugf("blockHashAndSendersMap[%d].BlockHash: %s", blockNumber, blockHashAndSenders.BlockHash)

		blockHashes[blockNumber] = intMaxTree.NewBlockHashLeaf(common.HexToHash("0x" + blockHashAndSenders.BlockHash))
		_, err = blockHashTree.AddLeaf(blockNumber, blockHashes[blockNumber])
		if err != nil {
			return errors.Join(ErrAddLeafToBlockHashTreeFail, err)
		}

		bbs.log.Debugf("blockHashTree %d root: %s", blockNumber, blockHashTree.GetRoot().String())

		merkleTreeHistory[blockNumber].BlockHashTree = new(intMaxTree.BlockHashTree).Set(blockHashTree)

		count := int0Key
		for i, sender := range blockHashAndSenders.Senders {
			if sender.PublicKey == defaultPublicKey || sender.PublicKey == dummyPublicKey {
				continue
			}

			bbs.log.Debugf("blockHashAndSendersMap[%d].Senders[%d]: %s", blockNumber, i, sender.PublicKey)

			count++

			var senderPublicKey *intMaxAcc.PublicKey
			senderPublicKey, err = intMaxAcc.NewPublicKeyFromAddressHex(sender.PublicKey)
			if err != nil {
				return errors.Join(ErrNewPublicKeyFromAddressHexFail, err)
			}

			if _, ok = accountTree.GetAccountID(senderPublicKey.BigInt()); ok {
				_, err = accountTree.Update(senderPublicKey.BigInt(), blockNumber)
				if err != nil {
					return errors.Join(ErrUpdateAccountTreeFail, err)
				}
			} else {
				_, err = accountTree.Insert(senderPublicKey.BigInt(), blockNumber)
				if err != nil {
					return errors.Join(ErrInsertAccountTreeFail, err)
				}
			}
		}

		bbs.log.Debugf("blockHashAndSendersMap[%d].Senders count: %d", blockNumber, count)

		merkleTreeHistory[blockNumber].AccountTree = new(intMaxTree.AccountTree).Set(accountTree)
	}

	bbs.log.Debugf("blockHashTree leaves: %v", blockHashTree.Leaves)

	for i, leaf := range blockHashTree.Leaves {
		bbs.log.Debugf("blockHashes[%d]: %x", i, leaf.Marshal())
	}

	var deposits []*mDBApp.Deposit
	deposits, err = db.ScanDeposits()
	if err != nil {
		return errors.Join(ErrScanDepositsFail, err)
	}

	blockNumber := uint32(int1Key)
	depositTreeRoot, _, _ := depositTree.GetCurrentRootCountAndSiblings()
	depositTreeRootHex := depositTreeRoot.Hex()[int2Key:]
	for blockHashAndSendersMap[blockNumber].DepositTreeRoot == depositTreeRootHex &&
		blockNumber <= lastProofGeneratedBlockNumber {
		merkleTreeHistory[blockNumber].DepositLeaves = make([]*intMaxTree.DepositLeaf, int0Key)
		blockNumber++
	}

	depositLeaves := make([]*intMaxTree.DepositLeaf, int0Key)
	for depositIndex, deposit := range deposits {
		depositLeaf := intMaxTree.DepositLeaf{
			RecipientSaltHash: deposit.RecipientSaltHash,
			TokenIndex:        deposit.TokenIndex,
			Amount:            deposit.Amount,
		}
		if deposit.DepositIndex == nil {
			// TODO: panic
			panic(fmt.Errorf("deposit index should not be nil"))
		}

		depositTreeRoot, err = depositTree.AddLeaf(uint32(depositIndex), depositLeaf.Hash())
		if err != nil {
			return errors.Join(ErrAddLeafToDepositTreeFail, err)
		}

		depositTreeRootHex = depositTreeRoot.Hex()[int2Key:]
		depositLeaves = append(depositLeaves, &depositLeaf)

		bbs.log.Debugf("depositTreeRoot: %s", blockHashAndSendersMap[blockNumber].DepositTreeRoot)

		for blockHashAndSendersMap[blockNumber].DepositTreeRoot == depositTreeRootHex &&
			blockNumber <= lastProofGeneratedBlockNumber {
			merkleTreeHistory[blockNumber].DepositLeaves = depositLeaves
			blockNumber++
		}
	}

	bbs.DepositTree = depositTree
	bbs.AccountTree = accountTree
	bbs.BlockTree = blockHashTree

	return nil
}

func (bbs *blockBuilderStorage) FetchUpdateWitness(
	db SQLDriverApp,
	publicKey *intMaxAcc.PublicKey,
	currentBlockNumber uint32,
	targetBlockNumber uint32,
	isPrevAccountTree bool,
) (*bbsTypes.UpdateWitness, error) {
	const (
		int1Key = 1
	)

	bbs.log.Debugf("FetchUpdateWitness currentBlockNumber: %d", currentBlockNumber)
	bbs.log.Debugf("FetchUpdateWitness targetBlockNumber: %d", targetBlockNumber)

	// request validity prover
	latestValidityProof, err := bbs.ValidityProofByBlockNumber(db, currentBlockNumber)
	if err != nil {
		return nil, errors.Join(ErrValidityProofByBlockNumberFail, err)
	}

	var blockMerkleProof *intMaxTree.PoseidonMerkleProof
	blockMerkleProof, err = bbs.BlockTreeProof(currentBlockNumber, targetBlockNumber)
	if err != nil {
		return nil, errors.Join(ErrBlockTreeProofFail, err)
	}

	var accountMembershipProof *intMaxTree.IndexedMembershipProof
	if isPrevAccountTree {
		bbs.log.Debugf("is PrevAccountTree %d", currentBlockNumber-int1Key)
		accountMembershipProof, err = bbs.GetAccountMembershipProof(currentBlockNumber-int1Key, publicKey.BigInt())
	} else {
		bbs.log.Debugf("is not PrevAccountTree %d", currentBlockNumber)
		accountMembershipProof, err = bbs.GetAccountMembershipProof(currentBlockNumber, publicKey.BigInt())
	}
	if err != nil {
		return nil, errors.Join(ErrGetAccountMembershipProofFail, err)
	}

	return &bbsTypes.UpdateWitness{
		ValidityProof:          *latestValidityProof,
		BlockMerkleProof:       *blockMerkleProof,
		AccountMembershipProof: accountMembershipProof,
	}, nil
}

func (bbs *blockBuilderStorage) ValidityProofByBlockNumber(db SQLDriverApp, blockNumber uint32) (*string, error) {
	if blockNumber == 0 {
		return nil, ErrGenesisValidityProof
	}

	blockContent, err := db.BlockContentByBlockNumber(blockNumber)
	if err != nil {
		bbs.log.WithError(err).Warnf("blockNumber (GetValidityProof): %d", blockNumber)
		return nil, errors.Join(ErrOfBlockContentByBlockNumber, err)
	}

	encodedValidityProof := base64.StdEncoding.EncodeToString(blockContent.ValidityProof)

	return &encodedValidityProof, nil
}

func (bbs *blockBuilderStorage) BlockTreeProof(
	rootBlockNumber, leafBlockNumber uint32,
) (*intMaxTree.PoseidonMerkleProof, error) {
	if rootBlockNumber < leafBlockNumber {
		const msg = "error of root block number less leaf block number: %d < %d"
		// TODO: panic
		panic(fmt.Errorf(msg, rootBlockNumber, leafBlockNumber))
	}

	blockHistory, ok := bbs.MerkleTreeHistory[rootBlockNumber]
	if !ok {
		const msg = "root block number %d not found (BlockTreeProof)"
		return nil, fmt.Errorf(msg, rootBlockNumber)
	}

	proof, _, err := blockHistory.BlockHashTree.Prove(leafBlockNumber)
	if err != nil {
		const msg = "leaf block number %d not found (BlockTreeProof)"
		return nil, fmt.Errorf(msg, leafBlockNumber)
	}

	return &proof, nil
}

func (bbs *blockBuilderStorage) GetAccountMembershipProof(
	blockNumber uint32,
	publicKey *big.Int,
) (*intMaxTree.IndexedMembershipProof, error) {
	blockHistory, ok := bbs.MerkleTreeHistory[blockNumber]
	if !ok {
		const msg = "current block number %d not found"
		return nil, fmt.Errorf(msg, blockNumber)
	}

	proof, _, err := blockHistory.AccountTree.ProveMembership(publicKey)
	if err != nil {
		const msg = "account membership proof error"
		return nil, errors.New(msg)
	}

	return proof, nil
}

func (bbs *blockBuilderStorage) GenerateBlock(
	blockContent *intMaxTypes.BlockContent,
	postedBlock *block_post_service.PostedBlock,
) (*bbsTypes.BlockWitness, error) {
	signature := bbsTypes.NewSignatureContentFromBlockContent(blockContent)
	publicKeys := make([]intMaxTypes.Uint256, len(blockContent.Senders))
	for i, sender := range blockContent.Senders {
		publicKey := new(intMaxTypes.Uint256).FromBigInt(sender.PublicKey.BigInt())
		publicKeys[i] = *publicKey
	}

	prevAccountTreeRoot := bbs.AccountTree.GetRoot()
	prevBlockTreeRoot := bbs.BlockTree.GetRoot()

	if signature.IsRegistrationBlock {
		accountMembershipProofs := make([]intMaxTree.IndexedMembershipProof, len(blockContent.Senders))
		for i, sender := range blockContent.Senders {
			accountMembershipProof, _, err := bbs.AccountTree.ProveMembership(sender.PublicKey.BigInt())
			if err != nil {
				return nil, errors.New("account membership proof error")
			}

			accountMembershipProofs[i] = *accountMembershipProof
		}

		blockWitness := &bbsTypes.BlockWitness{
			Block:                   postedBlock,
			Signature:               *signature,
			PublicKeys:              publicKeys,
			PrevAccountTreeRoot:     prevAccountTreeRoot,
			PrevBlockTreeRoot:       prevBlockTreeRoot,
			AccountIdPacked:         nil,
			AccountMerkleProofs:     nil,
			AccountMembershipProofs: &accountMembershipProofs,
		}

		return blockWitness, nil
	}

	accountMerkleProofs := make([]bbsTypes.AccountMerkleProof, len(blockContent.Senders))
	accountIDPackedBytes := make([]byte, bbsTypes.NumAccountIDPackedBytes)
	for i, sender := range blockContent.Senders {
		accountIDByte := make([]byte, bbsTypes.Int8Key)
		binary.BigEndian.PutUint64(accountIDByte, sender.AccountID)
		copy(accountIDPackedBytes[i/bbsTypes.Int8Key:i/bbsTypes.Int8Key+bbsTypes.Int5Key], accountIDByte[bbsTypes.Int8Key-bbsTypes.Int5Key:])
		accountMembershipProof, _, err := bbs.AccountTree.ProveMembership(sender.PublicKey.BigInt())
		if err != nil {
			return nil, errors.New("account membership proof error")
		}
		if !accountMembershipProof.IsIncluded {
			return nil, errors.New("account is not included")
		}

		accountMerkleProofs[i] = bbsTypes.AccountMerkleProof{
			MerkleProof: accountMembershipProof.LeafProof,
			Leaf:        accountMembershipProof.Leaf,
		}
	}

	accountIDPacked := new(bbsTypes.AccountIdPacked)
	accountIDPacked.FromBytes(accountIDPackedBytes)
	blockWitness := &bbsTypes.BlockWitness{
		Block:                   postedBlock,
		Signature:               *signature,
		PublicKeys:              publicKeys,
		PrevAccountTreeRoot:     prevAccountTreeRoot,
		PrevBlockTreeRoot:       prevBlockTreeRoot,
		AccountIdPacked:         accountIDPacked,
		AccountMerkleProofs:     &accountMerkleProofs,
		AccountMembershipProofs: nil,
	}

	return blockWitness, nil
}

func (bbs *blockBuilderStorage) FetchLastDepositIndex(db SQLDriverApp) (uint32, error) {
	return db.FetchLastDepositIndex()
}

func (bbs *blockBuilderStorage) LatestIntMaxBlockNumber() uint32 {
	return bbs.latestWitnessBlockNumber
}

func (bbs *blockBuilderStorage) LastPostedBlockNumber(db SQLDriverApp) (uint32, error) {
	return db.LastPostedBlockNumber()
}

func (bbs *blockBuilderStorage) EventBlockNumberByEventNameForValidityProver(
	db SQLDriverApp,
) (*mDBApp.EventBlockNumberForValidityProver, error) {
	return db.EventBlockNumberByEventNameForValidityProver(mDBApp.DepositsProcessedEvent)
}

func (bbs *blockBuilderStorage) LastSeenBlockPostedEventBlockNumber(db SQLDriverApp) (uint64, error) {
	event, err := db.EventBlockNumberByEventNameForValidityProver(mDBApp.BlockPostedEvent)
	if err != nil {
		return 0, err
	}

	return event.LastProcessedBlockNumber, err
}

func (bbs *blockBuilderStorage) SetLastSeenBlockPostedEventBlockNumber(db SQLDriverApp, blockNumber uint64) error {
	_, err := db.UpsertEventBlockNumberForValidityProver(mDBApp.BlockPostedEvent, blockNumber)

	return err
}

func (bbs *blockBuilderStorage) GetDepositInfoByHash(
	db SQLDriverApp,
	depositHash common.Hash,
) (*bbsTypes.DepositInfo, error) {
	depositLeafWithId, depositIndex, err := bbs.GetDepositLeafAndIndexByHash(db, depositHash)
	if err != nil {
		return nil, errors.Join(ErrGetDepositLeafAndIndexByHashFail, err)
	}

	depositInfo := bbsTypes.DepositInfo{
		DepositId:    depositLeafWithId.DepositId,
		DepositIndex: depositIndex,
		DepositLeaf:  depositLeafWithId.DepositLeaf,
	}
	if depositIndex != nil {
		var blockNumber uint32
		blockNumber, err = bbs.BlockNumberByDepositIndex(db, *depositIndex)
		if err != nil {
			return nil, errors.Join(ErrBlockNumberByDepositIndexFail, err)
		}

		var isSynchronizedDepositIndex bool
		isSynchronizedDepositIndex, err = bbs.IsSynchronizedDepositIndex(db, *depositIndex)
		if err != nil {
			return nil, errors.Join(ErrIsSynchronizedDepositIndexFail, err)
		}

		depositInfo.BlockNumber = &blockNumber
		depositInfo.IsSynchronized = isSynchronizedDepositIndex
	}

	return &depositInfo, nil
}

func (bbs *blockBuilderStorage) GetDepositLeafAndIndexByHash(
	db SQLDriverApp,
	depositHash common.Hash,
) (depositLeafWithId *bbsTypes.DepositLeafWithId, depositIndex *uint32, err error) {
	bbs.log.Debugf("GetDepositIndexByHash deposit hash: %s", depositHash.String())

	var deposit *mDBApp.Deposit
	deposit, err = db.DepositByDepositHash(depositHash)
	if err != nil {
		return nil, new(uint32), err
	}

	depositLeaf := intMaxTree.DepositLeaf{
		RecipientSaltHash: deposit.RecipientSaltHash,
		TokenIndex:        deposit.TokenIndex,
		Amount:            deposit.Amount,
	}

	return &bbsTypes.DepositLeafWithId{
		DepositId:   deposit.DepositID,
		DepositLeaf: &depositLeaf,
	}, deposit.DepositIndex, nil
}

// BlockNumberByDepositIndex
// TODO: refactor
func (bbs *blockBuilderStorage) BlockNumberByDepositIndex(
	db SQLDriverApp,
	depositIndex uint32,
) (uint32, error) {
	lastValidityWitness, err := bbs.LastValidityWitness(db)
	if err != nil {
		return 0, err
	}

	blockNumber := uint32(1)
	bbs.log.Debugf("lastValidityWitness.BlockWitness.Block.BlockNumber: %d", lastValidityWitness.BlockWitness.Block.BlockNumber)
	for ; blockNumber <= lastValidityWitness.BlockWitness.Block.BlockNumber; blockNumber++ {
		depositLeaves := bbs.MerkleTreeHistory[blockNumber].DepositLeaves
		bbs.log.Debugf("latest deposit index: %d", len(depositLeaves))
		if depositIndex >= uint32(len(depositLeaves)) {
			return 0, errors.New("deposit index is out of range")
		}
	}

	return blockNumber, nil
}

func (bbs *blockBuilderStorage) SetValidityWitness(blockNumber uint32, witness *bbsTypes.ValidityWitness) error {
	const int32Key = 32
	depositTree, err := intMaxTree.NewDepositTree(int32Key)
	if err != nil {
		return err
	}

	depositTreeRoot, _, _ := depositTree.GetCurrentRootCountAndSiblings()
	if depositTreeRoot != witness.BlockWitness.Block.DepositRoot {
		for i, deposit := range bbs.DepositLeaves {
			depositLeaf := intMaxTree.DepositLeaf{
				RecipientSaltHash: deposit.RecipientSaltHash,
				TokenIndex:        deposit.TokenIndex,
				Amount:            deposit.Amount,
			}

			_, err = depositTree.AddLeaf(uint32(i), depositLeaf)
			if err != nil {
				return err
			}

			depositTreeRoot, _, _ = depositTree.GetCurrentRootCountAndSiblings()
			bbs.log.Debugf("SetValidityWitness depositTreeRoot: %s", depositTreeRoot.String())
			if depositTreeRoot == witness.BlockWitness.Block.DepositRoot {
				break
			}
		}
	}

	bbs.log.Debugf("blockNumber: %d", blockNumber)
	bbs.log.Debugf("GetAccountMembershipProof root: %s", bbs.AccountTree.GetRoot().String())

	bbs.latestWitnessBlockNumber = blockNumber
	bbs.MerkleTreeHistory[blockNumber] = &bbsTypes.MerkleTrees{
		AccountTree:   new(intMaxTree.AccountTree).Set(bbs.AccountTree),
		BlockHashTree: new(intMaxTree.BlockHashTree).Set(bbs.BlockTree),
		DepositLeaves: depositTree.Leaves,
	}

	return nil
}

func (bbs *blockBuilderStorage) LastValidityWitness(db SQLDriverApp) (*bbsTypes.ValidityWitness, error) {
	return bbs.ValidityWitnessByBlockNumber(db, bbs.latestWitnessBlockNumber)
}

func (bbs *blockBuilderStorage) ValidityWitnessByBlockNumber(db SQLDriverApp, blockNumber uint32) (*bbsTypes.ValidityWitness, error) {
	if blockNumber == 0 {
		genesisValidityWitness := new(bbsTypes.ValidityWitness).Genesis()
		return genesisValidityWitness, nil
	}

	auxInfo, err := bbs.BlockAuxInfo(db, blockNumber)
	if err != nil {
		return nil, err
	}

	bbs.log.Debugf(
		"auxInfo.PostedBlock.BlockNumber (ValidityWitnessByBlockNumber): %d",
		auxInfo.PostedBlock.BlockNumber,
	)
	blockWitness, err := bbs.GenerateBlockWithTxTreeFromBlockContent(
		auxInfo.BlockContent,
		auxInfo.PostedBlock,
	)
	if err != nil {
		return nil, err
	}

	bbs.log.Debugf("blockNumber (ValidityWitnessByBlockNumber): %d", blockNumber)
	bbs.log.Debugf("blockWitness.Block.BlockNumber (ValidityWitnessByBlockNumber): %d", blockWitness.Block.BlockNumber)
	var validityWitness *bbsTypes.ValidityWitness
	validityWitness, err = bbs.CalculateValidityWitness(blockWitness)
	if err != nil {
		return nil, err
	}

	return validityWitness, nil
}

func (bbs *blockBuilderStorage) BlockAuxInfo(db SQLDriverApp, blockNumber uint32) (*bbsTypes.AuxInfo, error) {
	auxInfo, err := db.BlockContentByBlockNumber(blockNumber)
	if err != nil {
		return nil, errors.Join(ErrOfBlockContentByBlockNumber, err)
	}

	return bbsTypes.BlockAuxInfoFromBlockContent(bbs.log, auxInfo)
}

func (bbs *blockBuilderStorage) GenerateBlockWithTxTreeFromBlockContent(
	blockContent *intMaxTypes.BlockContent,
	postedBlock *block_post_service.PostedBlock,
) (*bbsTypes.BlockWitness, error) {
	const numOfSenders = 128
	if len(blockContent.Senders) > numOfSenders {
		return nil, errors.New("too many txs")
	}

	publicKeys := make([]intMaxTypes.Uint256, len(blockContent.Senders))
	for i, sender := range blockContent.Senders {
		publicKeys[i].FromBigInt(sender.PublicKey.BigInt())
	}

	dummyPublicKey := intMaxAcc.NewDummyPublicKey()
	for i := len(publicKeys); i < numOfSenders; i++ {
		publicKeys = append(publicKeys, *new(intMaxTypes.Uint256).FromBigInt(dummyPublicKey.BigInt()))
	}

	var (
		accountIDPacked         *bbsTypes.AccountIdPacked
		accountMerkleProofs     []bbsTypes.AccountMerkleProof
		accountMembershipProofs []intMaxTree.IndexedMembershipProof
	)
	isRegistrationBlock := blockContent.SenderType == "PUBLIC_KEY"
	if isRegistrationBlock {
		accountMembershipProofs = make([]intMaxTree.IndexedMembershipProof, len(publicKeys))
		bbs.log.Debugf("size of publicKeys: %d", len(publicKeys))
		for i, publicKey := range publicKeys {
			isDummy := publicKey.BigInt().Cmp(intMaxAcc.NewDummyPublicKey().BigInt()) == 0
			bbs.log.Debugf("isDummy: %v, ", isDummy)

			leaf, err := bbs.GetAccountTreeLeaf(publicKey.BigInt())
			if err != nil {
				if !errors.Is(err, ErrAccountTreeGetAccountID) {
					return nil, errors.Join(errors.New("account tree leaf error"), err)
				}
			}

			if !isDummy && leaf != nil {
				return nil, errors.New("account already exists")
			}

			var proof *intMaxTree.IndexedMembershipProof
			proof, err = bbs.GetAccountMembershipProof(postedBlock.BlockNumber, publicKey.BigInt())
			if err != nil {
				return nil, errors.Join(errors.New("account membership proof error"), err)
			}

			accountMembershipProofs[i] = *proof
		}
	} else {
		accountIDs := make([]uint64, len(publicKeys))
		accountMerkleProofs = make([]bbsTypes.AccountMerkleProof, len(publicKeys))
		for i, publicKey := range publicKeys {
			accountID, ok := bbs.AccountTree.GetAccountID(publicKey.BigInt())
			if !ok {
				return nil, errors.New("account id not found")
			}
			proof, err := bbs.ProveInclusion(accountID)
			if err != nil {
				return nil, errors.New("account inclusion proof error")
			}

			accountIDs[i] = accountID
			accountMerkleProofs[i] = bbsTypes.AccountMerkleProof{
				MerkleProof: proof.MerkleProof,
				Leaf:        proof.Leaf,
			}
		}

		accountIDPacked = new(bbsTypes.AccountIdPacked).Pack(accountIDs)
	}

	txTreeRoot := intMaxTypes.Bytes32{}
	txTreeRoot.FromBytes(blockContent.TxTreeRoot[:])
	signature := bbsTypes.NewSignatureContentFromBlockContent(blockContent)

	prevAccountTreeRoot := bbs.AccountTree.GetRoot()
	prevBlockTreeRoot := bbs.BlockTree.GetRoot()
	blockWitness := &bbsTypes.BlockWitness{
		Block:                   postedBlock,
		Signature:               *signature,
		PublicKeys:              publicKeys,
		PrevAccountTreeRoot:     prevAccountTreeRoot,
		PrevBlockTreeRoot:       prevBlockTreeRoot,
		AccountIdPacked:         accountIDPacked,
		AccountMerkleProofs:     &accountMerkleProofs,
		AccountMembershipProofs: &accountMembershipProofs,
	}

	return blockWitness, nil
}

func (bbs *blockBuilderStorage) AppendAccountTreeLeaf(
	sender *big.Int,
	lastBlockNumber uint32,
) (*intMaxTree.IndexedInsertionProof, error) {
	proof, err := bbs.AccountTree.Insert(sender, lastBlockNumber)
	if err != nil {
		// invalid block
		return nil, errors.Join(ErrAccountTreeInsert, err)
	}

	return proof, nil
}

func (bbs *blockBuilderStorage) GetAccountTreeLeaf(sender *big.Int) (*intMaxTree.IndexedMerkleLeaf, error) {
	accountID, ok := bbs.AccountTree.GetAccountID(sender)
	if !ok {
		return nil, ErrAccountTreeGetAccountID
	}
	prevLeaf := bbs.AccountTree.GetLeaf(accountID)

	return prevLeaf, nil
}

func (bbs *blockBuilderStorage) UpdateAccountTreeLeaf(
	sender *big.Int,
	lastBlockNumber uint32,
) (*intMaxTree.IndexedUpdateProof, error) {
	proof, err := bbs.AccountTree.Update(sender, lastBlockNumber)
	if err != nil {
		return nil, errors.Join(ErrAccountTreeUpdate, err)
	}

	return proof, nil
}

func (bbs *blockBuilderStorage) ProveInclusion(accountId uint64) (*bbsTypes.AccountMerkleProof, error) {
	leaf := bbs.AccountTree.GetLeaf(accountId)
	proof, _, err := bbs.AccountTree.Prove(accountId)
	if err != nil {
		return nil, err
	}

	return &bbsTypes.AccountMerkleProof{
		MerkleProof: *proof,
		Leaf:        *leaf,
	}, nil
}

func (bbs *blockBuilderStorage) BlockTreeRoot(blockNumber *uint32) (*intMaxGP.PoseidonHashOut, error) {
	if blockNumber == nil {
		return bbs.BlockTree.GetRoot(), nil
	}

	blockHistory, ok := bbs.MerkleTreeHistory[*blockNumber]
	if !ok {
		return nil, errors.New("block number not found")
	}

	return blockHistory.BlockHashTree.GetRoot(), nil
}

func (bbs *blockBuilderStorage) IsSynchronizedDepositIndex(db SQLDriverApp, depositIndex uint32) (bool, error) {
	lastGeneratedProofBlockNumber, err := bbs.LastGeneratedProofBlockNumber(db)
	if err != nil {
		return false, err
	}
	bbs.log.Debugf("lastPostedBlockNumber: %d", lastGeneratedProofBlockNumber)

	depositLeaves := bbs.MerkleTreeHistory[lastGeneratedProofBlockNumber].DepositLeaves
	bbs.log.Debugf("lastGeneratedProofBlockNumber (IsSynchronizedDepositIndex): %d", lastGeneratedProofBlockNumber)
	bbs.log.Debugf("latest deposit index: %d", len(depositLeaves))
	bbs.log.Debugf("depositIndex: %d", depositIndex)

	if depositIndex >= uint32(len(depositLeaves)) {
		return false, nil
	}

	return true, nil
}

func (bbs *blockBuilderStorage) LastGeneratedProofBlockNumber(db SQLDriverApp) (uint32, error) {
	lastValidityProof, err := db.LastBlockValidityProof()
	if err != nil {
		if errors.Is(err, errorsDB.ErrNotFound) {
			return 0, nil
		}

		return 0, err
	}

	return lastValidityProof.BlockNumber, nil
}

func (bbs *blockBuilderStorage) UpdateValidityWitness(
	blockContent *intMaxTypes.BlockContent,
	prevValidityWitness *bbsTypes.ValidityWitness,
) (*bbsTypes.ValidityWitness, error) {
	blockWitness, err := bbs.GenerateBlockWithTxTreeFromBlockContent(
		blockContent,
		prevValidityWitness.BlockWitness.Block,
	)
	if err != nil {
		panic(err)
	}

	bbs.log.Debugf("blockWitness.Block.BlockNumber (PostBlock): %d", blockWitness.Block.BlockNumber)
	latestIntMaxBlockNumber := bbs.LatestIntMaxBlockNumber()
	if blockWitness.Block.BlockNumber != latestIntMaxBlockNumber+1 {
		bbs.log.Debugf("latestIntMaxBlockNumber: %d", latestIntMaxBlockNumber)
		return nil, errors.New("block number is not equal to the last block number + 1")
	}

	var validityWitness *bbsTypes.ValidityWitness
	validityWitness, err = bbs.updateValidityWitnessWithConsistencyCheck(
		blockWitness,
		prevValidityWitness,
	)
	if err != nil {
		panic(err)
	}

	return validityWitness, nil
}

func (bbs *blockBuilderStorage) AccountTreeRoot() (*intMaxGP.PoseidonHashOut, error) {
	return bbs.AccountTree.GetRoot(), nil
}

func (bbs *blockBuilderStorage) AppendBlockTreeLeaf(
	block *block_post_service.PostedBlock,
) (blockNumber uint32, err error) {
	blockHashLeaf := intMaxTree.NewBlockHashLeaf(block.Hash())
	_, blockNumber, _ = bbs.BlockTree.GetCurrentRootCountAndSiblings()
	bbs.log.Debugf("next block number (AppendBlockTreeLeaf): %d", blockNumber)
	bbs.log.Debugf("block hashes (AppendBlockTreeLeaf): %v", bbs.BlockTree.Leaves)
	if blockNumber != block.BlockNumber {
		return 0, fmt.Errorf("block number is not equal to the current block number: %d != %d", blockNumber, block.BlockNumber)
	}
	bbs.log.Debugf("block hashes: %v", bbs.BlockTree.Leaves)
	bbs.log.Debugf("old block root: %s", bbs.BlockTree.GetRoot().String())

	newRoot, err := bbs.BlockTree.AddLeaf(blockNumber, blockHashLeaf)
	if err != nil {
		return 0, errors.Join(ErrBlockTreeAddLeaf, err)
	}
	bbs.log.Debugf("new block root: %s", newRoot.String())

	return blockNumber, nil
}

func (bbs *blockBuilderStorage) CreateBlockContent(
	db SQLDriverApp,
	postedBlock *block_post_service.PostedBlock,
	blockContent *intMaxTypes.BlockContent,
) (*mDBApp.BlockContentWithProof, error) {
	return db.CreateBlockContent(
		postedBlock,
		blockContent,
	)
}

func (bbs *blockBuilderStorage) BlockContentByTxRoot(db SQLDriverApp, txRoot common.Hash) (*mDBApp.BlockContentWithProof, error) {
	return db.BlockContentByTxRoot(txRoot)
}

func (bbs *blockBuilderStorage) UpdateDepositIndexByDepositHash(
	db SQLDriverApp,
	depositHash common.Hash,
	depositIndex uint32,
) error {
	err := db.UpdateDepositIndexByDepositHash(depositHash, depositIndex)
	if err != nil {
		return err
	}

	return nil
}

func (bbs *blockBuilderStorage) UpsertEventBlockNumberForValidityProver(
	db SQLDriverApp,
	eventName string,
	blockNumber uint64,
) (*mDBApp.EventBlockNumberForValidityProver, error) {
	return db.UpsertEventBlockNumberForValidityProver(eventName, blockNumber)
}

func (bbs *blockBuilderStorage) AppendDepositTreeLeaf(
	depositHash common.Hash,
	depositLeaf *intMaxTree.DepositLeaf,
) (root common.Hash, nextIndex uint32, err error) {
	_, nextIndex, _ = bbs.DepositTree.GetCurrentRootCountAndSiblings()
	bbs.DepositLeaves = append(bbs.DepositLeaves, depositLeaf)
	root, err = bbs.DepositTree.AddLeaf(nextIndex, depositHash)
	if err != nil {
		return [32]byte{}, 0, err
	}

	return root, nextIndex, nil
}

func (bbs *blockBuilderStorage) DepositTreeProof(
	blockNumber, depositIndex uint32,
) (*intMaxTree.KeccakMerkleProof, common.Hash, error) {
	bbs.log.Printf("blockNumber (DepositTreeProof): %d", blockNumber)

	depositLeaves := bbs.MerkleTreeHistory[blockNumber].DepositLeaves
	if depositIndex >= uint32(len(depositLeaves)) {
		return nil, common.Hash{}, errors.New("block number is out of range")
	}

	leaves := make([][32]byte, 0)
	for _, depositLeaf := range depositLeaves {
		leaves = append(leaves, depositLeaf.Hash())
	}
	proof, root, err := bbs.DepositTree.ComputeMerkleProof(depositIndex, leaves)
	if err != nil {
		return nil, common.Hash{}, errors.Join(ErrDepositTreeProof, err)
	}

	return proof, root, nil
}

func (bbs *blockBuilderStorage) LastDepositTreeRoot() (common.Hash, error) {
	root, _, _ := bbs.DepositTree.GetCurrentRootCountAndSiblings()

	return root, nil
}

func (bbs *blockBuilderStorage) NextAccountID() (uint64, error) {
	return uint64(bbs.AccountTree.Count()), nil
}

func (bbs *blockBuilderStorage) SetValidityProof(
	db SQLDriverApp,
	blockHash common.Hash,
	proof string,
) error {
	validityProof, err := base64.StdEncoding.DecodeString(proof)
	if err != nil {
		return err
	}

	_, err = db.CreateValidityProof(blockHash, validityProof)
	if err != nil {
		return err
	}

	return err
}

func (bbs *blockBuilderStorage) RegisterPublicKey(
	pk *intMaxAcc.PublicKey,
	_ uint32,
) (accountID uint64, err error) {
	const int0Key = 0

	publicKey := pk.BigInt()

	var proof *intMaxTree.IndexedMembershipProof
	proof, _, err = bbs.AccountTree.ProveMembership(publicKey)
	if err != nil {
		return int0Key, errors.Join(ErrProveMembershipFail, err)
	}

	if _, ok := bbs.AccountTree.GetAccountID(publicKey); ok {
		_, err = bbs.AccountTree.Update(publicKey, uint32(int0Key))
		if err != nil {
			return int0Key, errors.Join(ErrUpdateAccountFail, err)
		}

		return uint64(proof.LeafIndex), nil
	}

	var insertionProof *intMaxTree.IndexedInsertionProof
	insertionProof, err = bbs.AccountTree.Insert(publicKey, uint32(int0Key))
	if err != nil {
		return int0Key, errors.Join(ErrCreateAccountFail, err)
	}

	return uint64(insertionProof.Index), nil
}

func (bbs *blockBuilderStorage) PublicKeyByAccountID(accountID uint64) (pk *intMaxAcc.PublicKey, err error) {
	var accID uint256.Int
	accID.SetUint64(accountID)

	acc := bbs.AccountTree.GetLeaf(accountID)

	pk, err = new(intMaxAcc.PublicKey).SetBigInt(acc.Key)
	if err != nil {
		return nil, errors.Join(ErrDecodeHexToPublicKeyFail, err)
	}

	return pk, nil
}

func (bbs *blockBuilderStorage) AccountBySenderAddress(_ string) (*uint256.Int, error) {
	return nil, fmt.Errorf("AccountBySenderAddress not implemented")
}

func (bbs *blockBuilderStorage) CalculateValidityWitness(blockWitness *bbsTypes.BlockWitness) (*bbsTypes.ValidityWitness, error) {
	bbs.log.Debugf("---------------------- CalculateValidityWitness ----------------------")

	mainValidationPublicInputs := blockWitness.MainValidationPublicInputs()
	bbs.log.Debugf("mainValidationPublicInputs.BlockNumber: %d", mainValidationPublicInputs.BlockNumber)

	prevBlockNumber := mainValidationPublicInputs.BlockNumber - 1
	prevBlockTreeRoot, err := bbs.BlockTreeRoot(&prevBlockNumber)
	if err != nil {
		return nil, fmt.Errorf("block tree root error: %w", err)
	}

	return bbs.calculateValidityWitnessWithMerkleProofs(blockWitness, prevBlockTreeRoot)
}

func (bbs *blockBuilderStorage) CalculateValidityWitnessWithConsistencyCheck(
	blockWitness *bbsTypes.BlockWitness,
	prevValidityWitness *bbsTypes.ValidityWitness,
) (*bbsTypes.ValidityWitness, error) {
	bbs.log.Debugf("---------------------- calculateValidityWitnessWithConsistencyCheck ----------------------")

	prevPis := prevValidityWitness.ValidityPublicInputs(bbs.log)
	if blockWitness.Block.BlockNumber > prevPis.PublicState.BlockNumber+1 {
		bbs.log.Debugf(
			"blockWitness.Block.BlockNumber (generateValidityWitness): %d",
			blockWitness.Block.BlockNumber,
		)
		bbs.log.Debugf(
			"prevPis.PublicState.BlockNumber (generateValidityWitness): %d",
			prevPis.PublicState.BlockNumber,
		)
		return nil, errors.New("block number is not greater than the last block number")
	}

	accountTreeRoot, err := bbs.AccountTreeRoot()
	if err != nil {
		return nil, errors.New("account tree root error")
	}

	if !prevPis.PublicState.AccountTreeRoot.Equal(accountTreeRoot) {
		bbs.log.Debugf(
			"prevPis.PublicState.AccountTreeRoot is not the same with accountTreeRoot, %s != %s",
			prevPis.PublicState.AccountTreeRoot.String(),
			accountTreeRoot.String(),
		)
		return nil, errors.New("account tree root is not equal to the last account tree root")
	}

	var prevBlockTreeRoot *intMaxGP.PoseidonHashOut
	prevBlockTreeRoot, err = bbs.BlockTreeRoot(&prevPis.PublicState.BlockNumber)
	if err != nil {
		return nil, errors.New("block tree root error")
	}

	if prevPis.IsValidBlock {
		bbs.log.Debugf("block number %d is valid", prevPis.PublicState.BlockNumber+1)
	} else {
		bbs.log.Debugf("block number %d is invalid", prevPis.PublicState.BlockNumber+1)
	}
	bbs.log.Debugf("prevBlockTreeRoot: %s", prevBlockTreeRoot.String())

	if !prevPis.PublicState.BlockTreeRoot.Equal(prevBlockTreeRoot) {
		bbs.log.Debugf(
			"prevPis.PublicState.BlockTreeRoot is not the same with blockTreeRoot, %s != %s",
			prevPis.PublicState.BlockTreeRoot.String(),
			prevBlockTreeRoot.String(),
		)
		return nil, errors.New("block tree root is not equal to the last block tree root")
	}

	return bbs.calculateValidityWitnessWithMerkleProofs(blockWitness, prevBlockTreeRoot)
}

func (bbs *blockBuilderStorage) calculateValidityWitnessWithMerkleProofs(
	blockWitness *bbsTypes.BlockWitness,
	prevBlockTreeRoot *intMaxGP.PoseidonHashOut,
) (*bbsTypes.ValidityWitness, error) {
	blockMerkleProof, err := bbs.BlockTreeProof(blockWitness.Block.BlockNumber, blockWitness.Block.BlockNumber)
	if err != nil {
		return nil, errors.Join(ErrBlockTreeProve, err)
	}

	// Verify that the Merkle proof for the block hash tree is correct in its old state.
	defaultLeaf := new(intMaxTree.BlockHashLeaf).SetDefault()
	err = blockMerkleProof.Verify(
		defaultLeaf.Hash(),
		int(blockWitness.Block.BlockNumber),
		prevBlockTreeRoot,
	)
	if err != nil {
		panic(fmt.Errorf("old block merkle proof is invalid: %w", err))
	}

	blockHashLeaf := intMaxTree.NewBlockHashLeaf(blockWitness.Block.Hash())

	var newBlockTreeRoot *intMaxGP.PoseidonHashOut
	newBlockTreeRoot, err = bbs.BlockTreeRoot(&blockWitness.Block.BlockNumber)
	if err != nil {
		return nil, errors.New("block tree root error")
	}

	err = blockMerkleProof.Verify(
		blockHashLeaf.Hash(),
		int(blockWitness.Block.BlockNumber),
		newBlockTreeRoot,
	)
	if err != nil {
		panic("new block merkle proof is invalid")
	}

	senderLeaves := bbsTypes.GetSenderLeaves(blockWitness.PublicKeys, blockWitness.Signature.SenderFlag)

	blockPis := blockWitness.MainValidationPublicInputs()

	accountRegistrationProofsWitness := bbsTypes.AccountRegistrationProofs{
		IsValid: false,
		Proofs:  nil,
	}
	if blockPis.IsValid && blockPis.IsRegistrationBlock {
		accountRegistrationProofs := make([]intMaxTree.IndexedInsertionProof, 0)
		for _, senderLeaf := range senderLeaves {
			lastBlockNumber := uint32(0)
			if senderLeaf.IsValid {
				lastBlockNumber = blockPis.BlockNumber
			}

			var proof *intMaxTree.IndexedInsertionProof
			isDummy := senderLeaf.Sender.Cmp(intMaxAcc.NewDummyPublicKey().BigInt()) == 0
			if isDummy {
				proof = intMaxTree.NewDummyAccountRegistrationProof(intMaxTree.ACCOUNT_TREE_HEIGHT)
			} else {
				proof, err = bbs.AppendAccountTreeLeaf(senderLeaf.Sender, lastBlockNumber)
				if err != nil {
					return nil, errors.Join(ErrAppendAccountTreeLeaf, err)
				}
			}

			accountRegistrationProofs = append(accountRegistrationProofs, *proof)
		}

		accountRegistrationProofsWitness = bbsTypes.AccountRegistrationProofs{
			IsValid: true,
			Proofs:  accountRegistrationProofs,
		}
	}

	accountUpdateProofsWitness := bbsTypes.AccountUpdateProofs{
		IsValid: false,
		Proofs:  nil,
	}
	if blockPis.IsValid && !blockPis.IsRegistrationBlock {
		accountUpdateProofs := make([]intMaxTree.IndexedUpdateProof, 0, len(senderLeaves))
		for _, senderLeaf := range senderLeaves {
			var prevLeaf *intMaxTree.IndexedMerkleLeaf
			prevLeaf, err = bbs.GetAccountTreeLeaf(senderLeaf.Sender)
			if err != nil {
				return nil, errors.Join(ErrAccountTreeLeaf, err)
			}

			prevLastBlockNumber := uint32(prevLeaf.Value)
			lastBlockNumber := prevLastBlockNumber
			if senderLeaf.IsValid {
				lastBlockNumber = blockPis.BlockNumber
			}
			var proof *intMaxTree.IndexedUpdateProof
			proof, err = bbs.UpdateAccountTreeLeaf(senderLeaf.Sender, lastBlockNumber)
			if err != nil {
				return nil, errors.Join(ErrUpdateAccountTreeLeaf, err)
			}
			accountUpdateProofs = append(accountUpdateProofs, *proof)
		}

		accountUpdateProofsWitness = bbsTypes.AccountUpdateProofs{
			IsValid: true,
			Proofs:  accountUpdateProofs,
		}
	}

	bbs.log.Debugf("validity_witness prev_account_tree_root: %v", blockWitness.PrevAccountTreeRoot.String())
	bbs.log.Debugf("validity_witness accountRegistrationProofsWitness: %v", accountRegistrationProofsWitness)
	return &bbsTypes.ValidityWitness{
		BlockWitness: blockWitness,
		ValidityTransitionWitness: &bbsTypes.ValidityTransitionWitness{
			SenderLeaves:              senderLeaves,
			BlockMerkleProof:          *blockMerkleProof,
			AccountRegistrationProofs: accountRegistrationProofsWitness,
			AccountUpdateProofs:       accountUpdateProofsWitness,
		},
	}, nil
}

func (bbs *blockBuilderStorage) updateValidityWitnessWithConsistencyCheck(
	blockWitness *bbsTypes.BlockWitness,
	prevValidityWitness *bbsTypes.ValidityWitness,
) (*bbsTypes.ValidityWitness, error) {
	bbs.log.Debugf("---------------------- updateValidityWitnessWithConsistencyCheck ----------------------")
	// latestIntMaxBlockNumber := db.LatestIntMaxBlockNumber()
	prevPis := prevValidityWitness.ValidityPublicInputs(bbs.log)
	// blockWitness.Block.BlockNumber != latestIntMaxBlockNumber+1
	if blockWitness.Block.BlockNumber > prevPis.PublicState.BlockNumber+1 {
		bbs.log.Debugf("blockWitness.Block.BlockNumber (generateValidityWitness): %d", blockWitness.Block.BlockNumber)
		bbs.log.Debugf("prevPis.PublicState.BlockNumber (generateValidityWitness): %d", prevPis.PublicState.BlockNumber)
		return nil, errors.New("block number is not greater than the last block number")
	}

	accountTreeRoot, err := bbs.AccountTreeRoot()
	if err != nil {
		return nil, errors.New("account tree root error")
	}

	if !prevPis.PublicState.AccountTreeRoot.Equal(accountTreeRoot) {
		bbs.log.Debugf(
			"prevPis.PublicState.AccountTreeRoot is not the same with accountTreeRoot, %s != %s",
			prevPis.PublicState.AccountTreeRoot.String(),
			accountTreeRoot.String(),
		)
		return nil, errors.New("account tree root is not equal to the last account tree root")
	}

	var prevBlockTreeRoot *intMaxGP.PoseidonHashOut
	prevBlockTreeRoot, err = bbs.BlockTreeRoot(&prevPis.PublicState.BlockNumber)
	if err != nil {
		return nil, errors.New("block tree root error")
	}

	if prevPis.IsValidBlock {
		bbs.log.Debugf("block number %d is valid", prevPis.PublicState.BlockNumber+1)
	} else {
		bbs.log.Debugf("block number %d is invalid", prevPis.PublicState.BlockNumber+1)
	}

	bbs.log.Debugf("prevBlockTreeRoot: %s", prevBlockTreeRoot.String())

	if !prevPis.PublicState.BlockTreeRoot.Equal(prevBlockTreeRoot) {
		bbs.log.Debugf(
			"prevPis.PublicState.BlockTreeRoot is not the same with blockTreeRoot, %s != %s",
			prevPis.PublicState.BlockTreeRoot.String(),
			prevBlockTreeRoot.String(),
		)
		return nil, errors.New("block tree root is not equal to the last block tree root")
	}

	var addedBlockNumber uint32
	addedBlockNumber, err = bbs.AppendBlockTreeLeaf(blockWitness.Block)
	if err != nil {
		return nil, fmt.Errorf("append block tree leaf error: %w", err)
	}

	if addedBlockNumber != blockWitness.Block.BlockNumber {
		return nil, errors.New("block number is not equal to the added block number")
	}

	var validityWitness *bbsTypes.ValidityWitness
	validityWitness, err = bbs.calculateValidityWitnessWithMerkleProofs(blockWitness, prevBlockTreeRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate validity witness: %w", err)
	}

	bbs.log.Debugf(
		"blockWitness.Block.BlockNumber: %d",
		blockWitness.Block.BlockNumber,
	)
	bbs.log.Debugf(
		"validityWitness.BlockWitness.Block.BlockNumber: %d",
		validityWitness.BlockWitness.Block.BlockNumber,
	)
	bbs.log.Debugf(
		"SenderFlag 1: %v",
		validityWitness.BlockWitness.Signature.SenderFlag,
	)

	validityPis := validityWitness.ValidityPublicInputs(bbs.log)
	var encodedValidityPis []byte
	encodedValidityPis, err = json.Marshal(validityPis)
	if err != nil {
		panic(err)
	}

	bbs.log.Debugf("validityPis (PostBlock): %s", encodedValidityPis)
	bbs.log.Debugf("SetValidityWitness SenderFlag: %v", validityWitness.BlockWitness.Signature.SenderFlag)

	err = bbs.SetValidityWitness(blockWitness.Block.BlockNumber, validityWitness)
	if err != nil {
		panic(ErrSetValidityWitnessFail)
	}

	bbs.log.Debugf("post block #%d", validityWitness.BlockWitness.Block.BlockNumber)

	return validityWitness, nil
}
