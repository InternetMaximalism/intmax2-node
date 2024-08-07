package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	errorsB "intmax2-node/internal/blockchain/errors"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	modelsMW "intmax2-node/internal/mnemonic_wallet/models"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	"io/fs"
	"math/big"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/uuid"
	"github.com/holiman/uint256"
	"github.com/status-im/keycard-go/hexutils"
	bolt "go.etcd.io/bbolt"
)

const (
	bucket     = "transfers"
	int1024Key = 1024
)

type signaturesByLeafIndex struct {
	Sender    string
	TxHash    string
	Signature string
	LeafIndex uint64
	CreatedAt int64
}

type LeafsTree struct {
	TxTree                 *intMaxTree.TxTree
	TxRoot                 *intMaxTree.PoseidonHashOut
	Count                  uint64
	Siblings               []*intMaxTree.PoseidonHashOut
	SenderPublicKeys       []*intMaxAcc.PublicKey
	KeysOfSenderPublicKeys map[string]int
	SenderAccountIDs       []*uint256.Int
	SignaturesCounter      int
	Signatures             []*signaturesByLeafIndex
	SignaturesByLeafIndex  []*signaturesByLeafIndex
}

type kvInfo struct {
	Ctx                 context.Context
	CtxCancel           context.CancelFunc
	KvDB                *bolt.DB
	TransactionsCounter int32
	Delivered           bool
	Processing          bool
	Timestamp           *time.Time
	Receiver            chan func() error
	LeafsTreePublicKeys *LeafsTree
	LeafsTreeAccounts   *LeafsTree
}

type leafsOfHash struct {
	Sender    string
	AccountID *uint256.Int
	Index     uint64
}

type fileInfo struct {
	sync.Mutex
	kvInfo
	UsersCounter map[string]string
	Hashes       map[string]*leafsOfHash
}

type workerFileList struct {
	sync.Mutex
	CurrentDir  string
	CurrentFile *os.File
	FilesList   map[*os.File]*fileInfo
	Cleaner     chan func()
}

type TransactionHashesWithSenderAndFile struct {
	Sender string
	TxHash string
	File   *os.File
}

type transactionHashesList struct {
	sync.Mutex
	Hashes  map[string]*TransactionHashesWithSenderAndFile
	Cleaner chan func()
}

type worker struct {
	cfg        *configs.Config
	log        logger.Logger
	dbApp      SQLDriverApp
	files      *workerFileList
	trHashes   *transactionHashesList
	numWorkers int32
	maxWorkers int32
}

func New(cfg *configs.Config, log logger.Logger, dbApp SQLDriverApp) Worker {
	return &worker{
		cfg:   cfg,
		log:   log,
		dbApp: dbApp,
		files: &workerFileList{
			FilesList: make(map[*os.File]*fileInfo),
			Cleaner:   make(chan func(), int1024Key),
		},
		trHashes: &transactionHashesList{
			Hashes:  make(map[string]*TransactionHashesWithSenderAndFile),
			Cleaner: make(chan func(), int1024Key),
		},
		maxWorkers: cfg.Worker.MaxCounter,
	}
}

func (w *worker) Init() (err error) {
	err = w.newTempDir()
	if err != nil {
		return errors.Join(ErrCreateNewTempDirFail, err)
	}

	err = w.newTempFile(w.files.CurrentDir)
	if err != nil {
		return errors.Join(ErrCreateNewTempFileFail, err)
	}

	return nil
}

func (w *worker) newTempDir() (err error) {
	w.files.Lock()
	defer w.files.Unlock()

	const (
		zeroPattern = "*"
		emptyKey    = ""
	)
	w.cfg.Worker.Path = strings.TrimSpace(w.cfg.Worker.Path)
	if w.cfg.Worker.Path == emptyKey {
		w.cfg.Worker.Path, err = os.MkdirTemp(emptyKey, zeroPattern)
		if err != nil {
			return errors.Join(ErrMkdirTempFail, err)
		}
	}

	if w.cfg.Worker.PathCleanInStart {
		err = os.RemoveAll(w.cfg.Worker.Path)
		if err != nil {
			return errors.Join(ErrRemoveAllFail, err)
		}
	}

	w.cfg.Worker.ID = strings.TrimSpace(w.cfg.Worker.ID)
	if w.cfg.Worker.ID == emptyKey {
		w.cfg.Worker.ID = uuid.New().String()
	}

	const maskMkdir = "%s%s%s"
	path := fmt.Sprintf(
		maskMkdir, w.cfg.Worker.Path, string(os.PathSeparator), w.cfg.Worker.ID,
	)

	err = os.RemoveAll(path)
	if err != nil {
		return errors.Join(ErrRemoveAllFail, err)
	}

	err = os.MkdirAll(path, fs.ModePerm)
	if err != nil {
		return errors.Join(ErrMkdirFail, err)
	}

	w.files.CurrentDir = path

	return nil
}

func (w *worker) newTempFile(dir string) error {
	w.files.Lock()
	defer w.files.Unlock()

	const zeroPattern = "*"

	currentFile, err := os.CreateTemp(dir, zeroPattern)
	if err != nil {
		return errors.Join(ErrCreateTempFail, err)
	}

	var kv *bolt.DB
	kv, err = w.kvStore(currentFile.Name())
	if err != nil {
		return errors.Join(ErrKVStoreFail, err)
	}

	ctxFileInfo, cancelFileInfo := context.WithCancel(context.Background())
	w.files.FilesList[currentFile] = &fileInfo{
		kvInfo: kvInfo{
			Ctx:       ctxFileInfo,
			CtxCancel: cancelFileInfo,
			KvDB:      kv,
			Receiver:  make(chan func() error, int1024Key),
		},
		UsersCounter: make(map[string]string),
		Hashes:       make(map[string]*leafsOfHash),
	}

	if w.files.CurrentFile != nil {
		if _, ok := w.files.FilesList[w.files.CurrentFile]; ok {
			tm := time.Now().UTC()
			w.files.FilesList[w.files.CurrentFile].Timestamp = &tm
		}
	}
	w.files.CurrentFile = currentFile

	go func() {
		for {
			select {
			case <-w.files.FilesList[currentFile].Ctx.Done():
				w.files.Lock()
				w.files.FilesList[currentFile].Delivered = true
				_ = w.files.FilesList[currentFile].KvDB.Close()
				w.files.Unlock()
				return
			case fn := <-w.files.FilesList[currentFile].Receiver:
				errTx := fn()
				if errTx != nil {
					w.log.Errorf("%+v", errTx)
				}
			}
		}
	}()

	w.files.FilesList[currentFile].Receiver <- func() error {
		var tx *bolt.Tx
		tx, err = kv.Begin(true)
		if err != nil {
			return errors.Join(ErrTxBeginKVStoreFail, err)
		}
		defer func() {
			_ = tx.Rollback()
		}()

		_, err = tx.CreateBucket([]byte(bucket))
		if err != nil {
			return errors.Join(ErrCreateBucketKVStoreFail, err)
		}

		err = tx.Commit()
		if err != nil {
			return errors.Join(ErrTxCommitKVStoreFail, err)
		}

		return nil
	}

	return nil
}

func (w *worker) CurrentDir() string {
	return w.files.CurrentDir
}

func (w *worker) CurrentFileName() string {
	return w.files.CurrentFile.Name()
}

func (w *worker) AvailableFiles() (list []*os.File, err error) {
	for key := range w.files.FilesList {
		cond1 := w.files.CurrentFile.Name() != key.Name()
		cond2 := atomic.LoadInt32(&w.files.FilesList[key].TransactionsCounter) == 0
		if cond1 && cond2 && !w.files.FilesList[key].Processing {
			if !w.files.FilesList[key].Delivered {
				err = w.leafsProcessing(key)
				if err != nil {
					return nil, errors.Join(ErrLeafsProcessing, err)
				}

				w.files.FilesList[key].CtxCancel()
				continue
			}
			list = append(list, key)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		b, _ := list[i].Stat()
		b2, _ := list[j].Stat()
		return b.ModTime().Before(b2.ModTime())
	})

	return list, nil
}

func (w *worker) TxTreeByAvailableFile(sf *TransactionHashesWithSenderAndFile) (txTreeRoot *TxTree, err error) {
	f, ok := w.files.FilesList[sf.File]
	if !ok {
		// transfersHash not found
		return nil, ErrTxTreeByAvailableFileFail
	}

	switch {
	case
		w.files.CurrentFile.Name() == sf.File.Name(),
		atomic.LoadInt32(&f.TransactionsCounter) != 0,
		!f.Delivered:
		// transfersHash exists, tx tree not found
		return nil, ErrTxTreeNotFound
	case
		w.files.CurrentFile.Name() != sf.File.Name() &&
			atomic.LoadInt32(&f.TransactionsCounter) == 0 &&
			f.Delivered &&
			f.Timestamp != nil && f.Timestamp.UTC().Add(
			w.cfg.Worker.TimeoutForSignaturesAvailableFiles,
		).UnixNano() <= time.Now().UTC().UnixNano():
		for {
			if !f.Processing {
				// transfersHash exists, tx tree exists, signature collection for tx tree completed
				return nil, ErrTxTreeSignatureCollectionComplete
			}
		}
	case
		w.files.CurrentFile.Name() != sf.File.Name() &&
			atomic.LoadInt32(&f.TransactionsCounter) == 0 &&
			f.Delivered &&
			f.Timestamp != nil && f.Timestamp.UTC().Add(
			w.cfg.Worker.TimeoutForSignaturesAvailableFiles,
		).UnixNano() > time.Now().UTC().UnixNano():
		for {
			if !f.Processing {
				break
			}
		}
	}

	var (
		siblings []*intMaxTree.PoseidonHashOut
		root     intMaxTree.PoseidonHashOut
	)

	if f.Hashes[sf.TxHash].AccountID != nil {
		siblings, root, err = f.LeafsTreeAccounts.TxTree.ComputeMerkleProof(f.Hashes[sf.TxHash].Index)
		if err != nil {
			return nil, errors.Join(ErrTxTreeComputeMerkleProofFail, err)
		}

		txTreeRoot = &TxTree{
			RootHash:         &root,
			Siblings:         siblings,
			SenderPublicKeys: f.LeafsTreeAccounts.SenderPublicKeys,
		}

		return txTreeRoot, nil
	}

	siblings, root, err = f.LeafsTreePublicKeys.TxTree.ComputeMerkleProof(f.Hashes[sf.TxHash].Index)
	if err != nil {
		return nil, errors.Join(ErrTxTreeComputeMerkleProofFail, err)
	}

	txTreeRoot = &TxTree{
		RootHash:         &root,
		Siblings:         siblings,
		SenderPublicKeys: f.LeafsTreePublicKeys.SenderPublicKeys,
	}

	return txTreeRoot, err
}

func (w *worker) ExistsTxTreeRoot(
	txTreeRoot string,
) error {
	for key := range w.files.FilesList {
		fmt.Printf("key: %v\n", key)

		if w.files.FilesList[key].LeafsTree == nil {
			continue
		}
		if w.files.FilesList[key].LeafsTree.TxRoot == nil {
			continue
		}

		actualTxTreeRoot, _, _ := w.files.FilesList[key].LeafsTree.TxTree.GetCurrentRootCountAndSiblings()
		fmt.Printf("actualTxTreeRoot: %v\n", actualTxTreeRoot.String())
		if actualTxTreeRoot.String() == txTreeRoot {
			return nil
		}
	}

	var ErrTxTreeRootNotFound = errors.New("tx tree root not found")

	return ErrTxTreeRootNotFound
}

func (w *worker) Start(
	ctx context.Context,
	tickerCurrentFile, tickerSignaturesAvailableFiles *time.Ticker,
) error {
	for {
		select {
		case <-ctx.Done():
			tickerCurrentFile.Stop()
			return nil
		case <-tickerCurrentFile.C:
			st, err := w.files.CurrentFile.Stat()
			if err != nil {
				return errors.Join(ErrStatCurrentFileFail, err)
			}

			// cond1 - current file lifetime expired
			cond1 := st.ModTime().UTC().Add(w.cfg.Worker.CurrentFileLifetime).UnixNano()-time.Now().UTC().UnixNano() <= 0
			// cond2 - the number of users exceeded the limit
			cond2 := len(w.files.FilesList[w.files.CurrentFile].UsersCounter) > w.cfg.Worker.MaxCounterOfUsers
			if cond1 || cond2 {
				err = w.newTempFile(w.files.CurrentDir)
				if err != nil {
					return errors.Join(ErrCreateNewTempFileFail, err)
				}
			}
		case <-tickerSignaturesAvailableFiles.C:
			list, err := w.AvailableFiles()
			if err != nil {
				return errors.Join(ErrAvailableFilesProcessing, err)
			}

			for key := range list {
				// cond1 - all transactions are processed
				cond1 := w.files.FilesList[list[key]].Delivered
				// cond2 - transaction collection for tx tree completed
				cond2 := w.files.FilesList[list[key]].Timestamp != nil &&
					w.files.FilesList[list[key]].Timestamp.UTC().Add(
						w.cfg.Worker.TimeoutForSignaturesAvailableFiles,
					).UnixNano() < time.Now().UTC().UnixNano()
				if cond1 && cond2 {
					if atomic.LoadInt32(&w.numWorkers) < w.maxWorkers {
						// Change status to processing
						w.files.FilesList[list[key]].Processing = true
						atomic.AddInt32(&w.numWorkers, 1)
						go func(f *os.File) {
							if err = w.postProcessing(ctx, f); err != nil {
								const msg = "failed to apply post processing"
								w.log.WithError(err).Errorf(msg)
							}
						}(list[key])
					}
				}
			}
		case f := <-w.trHashes.Cleaner:
			f()
		case f := <-w.files.Cleaner:
			f()
		}
	}
}

func (w *worker) Receiver(input *ReceiverWorker) error {
	if input == nil {
		return ErrReceiverWorkerEmpty
	}

	transfersHashBytes, err := hexutil.Decode(input.TransfersHash)
	if err != nil {
		return errors.Join(ErrHexDecodeFail, err)
	}
	transfersHash := new(intMaxGP.PoseidonHashOut)
	err = transfersHash.Unmarshal(transfersHashBytes)
	if err != nil {
		return errors.Join(ErrUnmarshalFail, err)
	}

	var currTx *intMaxTypes.Tx
	currTx, err = intMaxTypes.NewTx(
		transfersHash,
		input.Nonce,
	)
	if err != nil {
		return errors.Join(ErrNewTxFail, err)
	}

	w.trHashes.Lock()
	defer w.trHashes.Unlock()

	_, ok := w.trHashes.Hashes[currTx.Hash().String()]
	if ok {
		return ErrReceiverWorkerDuplicate
	}

	input.TxHash = currTx

	w.trHashes.Hashes[currTx.Hash().String()] = &TransactionHashesWithSenderAndFile{
		Sender: input.Sender,
		TxHash: currTx.Hash().String(),
		File:   w.files.CurrentFile,
	}

	err = w.registerReceiver(input)
	if err != nil {
		return errors.Join(ErrRegisterReceiverFail, err)
	}

	return nil
}

func (w *worker) registerReceiver(input *ReceiverWorker) (err error) {
	if input == nil {
		return ErrReceiverWorkerEmpty
	}

	current := w.files.CurrentFile

	w.files.Lock()
	atomic.AddInt32(&w.files.FilesList[current].TransactionsCounter, 1)
	w.files.FilesList[current].UsersCounter[input.Sender] = input.Sender
	w.files.FilesList[current].Hashes[input.TxHash.Hash().String()] = nil
	w.files.Unlock()

	w.files.FilesList[current].Receiver <- func() error {
		w.files.Lock()
		defer func() {
			atomic.AddInt32(&w.files.FilesList[current].TransactionsCounter, -1)
			w.files.Unlock()
		}()

		var tx *bolt.Tx
		tx, err = w.files.FilesList[current].KvDB.Begin(true)
		if err != nil {
			return errors.Join(ErrTxBeginKVStoreFail, err)
		}
		defer func() {
			_ = tx.Rollback()
		}()

		b := tx.Bucket([]byte(bucket))

		var (
			s    []byte
			info SenderInfo
		)
		s = b.Get([]byte(input.Sender))
		if s != nil {
			err = json.Unmarshal(s, &info)
			if err != nil {
				return errors.Join(ErrUnmarshalFail, err)
			}
		} else {
			info = SenderInfo{
				Sender:  &intMaxTypes.Sender{},
				TxsList: make(map[string]*ReceiverWorker),
			}
		}

		info.TxsList[input.TxHash.Hash().String()] = input

		var bST []byte
		bST, err = json.Marshal(&info)
		if err != nil {
			return errors.Join(ErrMarshalFail, err)
		}

		err = b.Put([]byte(input.Sender), bST)
		if err != nil {
			return errors.Join(ErrPutBucketKVStoreFail, err)
		}

		err = tx.Commit()
		if err != nil {
			return errors.Join(ErrTxCommitKVStoreFail, err)
		}

		return nil
	}

	return nil
}

func (w *worker) leafsProcessing(f *os.File) (err error) {
	defer atomic.AddInt32(&w.numWorkers, -1)

	var txTreePublicKeys *intMaxTree.TxTree
	txTreePublicKeys, err = intMaxTree.NewTxTree(
		intMaxTree.TX_TREE_HEIGHT, []*intMaxTypes.Tx{}, new(intMaxTypes.PoseidonHashOut).SetZero(),
	)
	if err != nil {
		return errors.Join(ErrNewTxTreeFail, err)
	}

	var txTreeAccountIDs *intMaxTree.TxTree
	txTreeAccountIDs, err = intMaxTree.NewTxTree(
		intMaxTree.TX_TREE_HEIGHT, []*intMaxTypes.Tx{}, new(intMaxTypes.PoseidonHashOut).SetZero(),
	)
	if err != nil {
		return errors.Join(ErrNewTxTreeFail, err)
	}

	spKeys := make(map[string]*intMaxAcc.PublicKey)

	spKeysByAccIDs := make(map[string]*intMaxAcc.PublicKey)
	accIDsByAccIDs := make(map[string]*uint256.Int)

	var numberPublicKeys, numberAccountIDs int
	err = w.files.FilesList[f].KvDB.View(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var info SenderInfo
			err = json.Unmarshal(v, &info)
			if err != nil {
				err = errors.Join(ErrUnmarshalFail, err)
				return err
			}

			for key := range info.TxsList {
				var accID uint256.Int
				err = w.dbApp.Exec(w.files.FilesList[f].Ctx, &accID, func(d interface{}, in interface{}) error {
					q := d.(SQLDriverApp)

					ai := block_post_service.NewAccountInfo(q)
					var accIDInfo *uint256.Int
					accIDInfo, err = ai.AccountBySenderAddress(info.TxsList[key].Sender)
					if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
						return err
					}
					if errors.Is(err, errorsDB.ErrNotFound) {
						return nil
					}

					if inV, ok := in.(*uint256.Int); ok {
						*inV = *accIDInfo
					} else {
						const msg = "failed to convert of account ID from uint256.Int"
						return fmt.Errorf(msg)
					}

					return nil
				})
				if err != nil {
					const msg = "failed to get of account ID with DBApp"
					return fmt.Errorf(msg)
				}

				lfh := leafsOfHash{
					Sender: info.TxsList[key].Sender,
				}

				var isAcc bool
				if accID.ToBig().Cmp(new(big.Int)) == 1 {
					isAcc = true
				}

				if isAcc {
					lfh.Index = uint64(numberAccountIDs)
					var lhfAccID uint256.Int
					_ = lhfAccID.SetFromBig(accID.ToBig())
					lfh.AccountID = &lhfAccID
					_, err = txTreeAccountIDs.AddLeaf(lfh.Index, info.TxsList[key].TxHash)
					if err != nil {
						return errors.Join(ErrAddLeafIntoTxTreeFail, err)
					}

					if _, ok := spKeysByAccIDs[info.TxsList[key].Sender]; !ok {
						var publicKey *intMaxAcc.PublicKey
						publicKey, err = intMaxAcc.NewPublicKeyFromAddressHex(info.TxsList[key].Sender)
						if err != nil {
							return errors.Join(ErrNewPublicKeyFromAddressHexFail, err)
						}
						spKeysByAccIDs[info.TxsList[key].Sender] = publicKey
						accIDsByAccIDs[info.TxsList[key].Sender] = lfh.AccountID
					}
				} else {
					lfh.Index = uint64(numberPublicKeys)
					_, err = txTreePublicKeys.AddLeaf(lfh.Index, info.TxsList[key].TxHash)
					if err != nil {
						return errors.Join(ErrAddLeafIntoTxTreeFail, err)
					}

					if _, ok := spKeys[info.TxsList[key].Sender]; !ok {
						var publicKey *intMaxAcc.PublicKey
						publicKey, err = intMaxAcc.NewPublicKeyFromAddressHex(info.TxsList[key].Sender)
						if err != nil {
							return errors.Join(ErrNewPublicKeyFromAddressHexFail, err)
						}
						spKeys[info.TxsList[key].Sender] = publicKey
					}
				}

				w.files.Lock()
				w.files.FilesList[f].Hashes[info.TxsList[key].TxHash.Hash().String()] = &lfh
				w.files.Unlock()

				if isAcc {
					numberAccountIDs++
				} else {
					numberPublicKeys++
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if numberPublicKeys > 0 {
		var (
			senderPublicKeys []*intMaxAcc.PublicKey
		)
		for key := range spKeys {
			senderPublicKeys = append(senderPublicKeys, spKeys[key])
		}

		// Sort by x-coordinate of public key
		sort.Slice(senderPublicKeys, func(i, j int) bool {
			return senderPublicKeys[i].Pk.X.Cmp(&senderPublicKeys[j].Pk.X) > 0
		})

		keysOfSenderPublicKeys := make(map[string]int)
		for key := range senderPublicKeys {
			keysOfSenderPublicKeys[senderPublicKeys[key].ToAddress().String()] = key
		}

		txRoot, count, sb := txTreePublicKeys.GetCurrentRootCountAndSiblings()

		w.files.FilesList[f].Lock()
		w.files.FilesList[f].LeafsTreePublicKeys = &LeafsTree{
			TxTree:                 txTreePublicKeys,
			TxRoot:                 &txRoot,
			Count:                  count,
			Siblings:               sb,
			SenderPublicKeys:       senderPublicKeys,
			KeysOfSenderPublicKeys: keysOfSenderPublicKeys,
			Signatures:             make([]*signaturesByLeafIndex, len(senderPublicKeys)),
			SignaturesByLeafIndex:  make([]*signaturesByLeafIndex, numberPublicKeys),
		}
		w.files.FilesList[f].Unlock()
	}

	if numberAccountIDs > 0 {
		var (
			senderPublicKeys []*intMaxAcc.PublicKey
			senderAccountIDs []*uint256.Int
		)
		for key := range spKeysByAccIDs {
			senderPublicKeys = append(senderPublicKeys, spKeysByAccIDs[key])
			senderAccountIDs = append(senderAccountIDs, accIDsByAccIDs[key])
		}

		// Sort by x-coordinate of public key
		sort.Slice(senderPublicKeys, func(i, j int) bool {
			return senderPublicKeys[i].Pk.X.Cmp(&senderPublicKeys[j].Pk.X) > 0
		})

		keysOfSenderPublicKeys := make(map[string]int)
		for key := range senderPublicKeys {
			keysOfSenderPublicKeys[senderPublicKeys[key].ToAddress().String()] = key
		}

		txRoot, count, sb := txTreePublicKeys.GetCurrentRootCountAndSiblings()

		w.files.FilesList[f].Lock()
		w.files.FilesList[f].LeafsTreeAccounts = &LeafsTree{
			TxTree:                 txTreePublicKeys,
			TxRoot:                 &txRoot,
			Count:                  count,
			Siblings:               sb,
			SenderPublicKeys:       senderPublicKeys,
			KeysOfSenderPublicKeys: keysOfSenderPublicKeys,
			SenderAccountIDs:       senderAccountIDs,
			Signatures:             make([]*signaturesByLeafIndex, len(senderPublicKeys)),
			SignaturesByLeafIndex:  make([]*signaturesByLeafIndex, numberAccountIDs),
		}
		w.files.FilesList[f].Unlock()
	}

	return nil
}

func (w *worker) postProcessing(ctx context.Context, f *os.File) (err error) {
	defer atomic.AddInt32(&w.numWorkers, -1)

	if len(w.files.FilesList[f].Hashes) == 0 {
		w.files.Lock()
		defer w.files.Unlock()
		delete(w.files.FilesList, f)
		w.files.Cleaner <- func() {
			_ = os.Remove(f.Name())
		}
		return nil
	}

	defer func() {
		w.files.Lock()
		defer w.files.Unlock()
		for key := range w.files.FilesList[f].Hashes {
			delete(w.files.FilesList[f].Hashes, key)
			w.trHashes.Cleaner <- func() {
				w.trHashes.Lock()
				defer w.trHashes.Unlock()
				delete(w.trHashes.Hashes, key)
			}
		}
		w.files.FilesList[f].Processing = false
	}()

	const int0Key = 0
	if (w.files.FilesList[f].LeafsTreePublicKeys == nil ||
		w.files.FilesList[f].LeafsTreePublicKeys.SignaturesCounter <= int0Key) &&
		(w.files.FilesList[f].LeafsTreeAccounts == nil ||
			w.files.FilesList[f].LeafsTreeAccounts.SignaturesCounter <= int0Key) {
		return nil
	}

	var mw *modelsMW.Wallet
	mw, err = mnemonic_wallet.New().WalletFromPrivateKeyHex(w.cfg.Blockchain.BuilderPrivateKeyHex)
	if err != nil {
		return errors.Join(errorsB.ErrWalletAddressNotRecognized, err)
	}

	err = w.dbApp.Exec(ctx, nil, func(d interface{}, _ interface{}) (err error) {
		q := d.(SQLDriverApp)

		funcLFT := func(block *mDBApp.Block, lft *LeafsTree) error {
			for index := range lft.SignaturesByLeafIndex {
				var sign *mDBApp.Signature
				if lft.SignaturesByLeafIndex[index].Signature != "" {
					sign, err = q.CreateSignature(
						lft.SignaturesByLeafIndex[index].Signature,
						block.ProposalBlockID,
					)
					if err != nil {
						return errors.Join(ErrCreateSignatureFail, err)
					}
				}

				var publicKey *intMaxAcc.PublicKey
				publicKey, err = intMaxAcc.NewPublicKeyFromAddressHex(
					lft.SignaturesByLeafIndex[index].Sender,
				)
				if err != nil {
					err = errors.Join(ErrNewPublicKeyFromAddressHexFail, err)
					return err
				}

				var txTreeIndex uint256.Int
				_ = txTreeIndex.SetUint64(lft.SignaturesByLeafIndex[index].LeafIndex)

				var cmp ComputeMerkleProof
				cmp.Siblings, cmp.Root, err = lft.TxTree.ComputeMerkleProof(
					lft.SignaturesByLeafIndex[index].LeafIndex,
				)
				if err != nil {
					return errors.Join(ErrTxTreeComputeMerkleProofFail, err)
				}

				var bCMP []byte
				bCMP, err = json.Marshal(&cmp)
				if err != nil {
					return errors.Join(ErrMarshalFail, err)
				}

				var signatureID string
				if sign != nil {
					signatureID = sign.SignatureID
				}
				_, err = q.CreateTxMerkleProofs(
					publicKey.String(),
					lft.SignaturesByLeafIndex[index].TxHash,
					signatureID,
					&txTreeIndex,
					bCMP,
					lft.TxRoot.String(),
					block.ProposalBlockID,
				)
				if err != nil {
					return errors.Join(ErrCreateTxMerkleProofsFail, err)
				}
			}

			return nil
		}

		var lft *LeafsTree
		if w.files.FilesList[f].LeafsTreePublicKeys != nil &&
			w.files.FilesList[f].LeafsTreePublicKeys.SignaturesCounter > int0Key {
			lft = w.files.FilesList[f].LeafsTreePublicKeys

			var bytesLfsTree []byte
			bytesLfsTree, err = json.Marshal(&lft)
			if err != nil {
				return errors.Join(ErrMarshalFail, err)
			}

			signatures := make([]string, len(lft.SenderPublicKeys))
			for indexSPK := range lft.SenderPublicKeys {
				signatures[indexSPK] = lft.Signatures[indexSPK].Signature
			}

			var bc *intMaxTypes.BlockContent
			bc, err = block_post_service.MakeRegistrationBlock(
				*lft.TxRoot,
				lft.SenderPublicKeys,
				signatures,
			)
			if err != nil {
				return errors.Join(ErrMakeRegistrationBlockFail, err)
			}

			senders := make([]intMaxTypes.ColumnSender, 0)
			for i := range bc.Senders {
				senders = append(senders, intMaxTypes.ColumnSender{
					PublicKey: bc.Senders[i].PublicKey.ToAddress().String(),
					AccountID: bc.Senders[i].AccountID,
					IsSigned:  bc.Senders[i].IsSigned,
				})
			}

			var block *mDBApp.Block
			block, err = q.CreateBlock(
				mw.IntMaxPublicKey,
				hexutils.BytesToHex(bc.TxTreeRoot.Marshal()),
				hexutils.BytesToHex(bc.AggregatedSignature.Marshal()),
				hexutils.BytesToHex(bc.AggregatedPublicKey.Marshal()),
				senders,
				mDBApp.ST_PUBLIC_KEY,
				bytesLfsTree,
			)
			if err != nil {
				return errors.Join(ErrCreateBlockFail, err)
			}

			return funcLFT(block, lft)
		}

		if w.files.FilesList[f].LeafsTreeAccounts != nil &&
			w.files.FilesList[f].LeafsTreeAccounts.SignaturesCounter > int0Key {
			lft = w.files.FilesList[f].LeafsTreeAccounts

			var bytesLfsTree []byte
			bytesLfsTree, err = json.Marshal(&lft)
			if err != nil {
				return errors.Join(ErrMarshalFail, err)
			}

			signatures := make([]string, len(lft.SenderPublicKeys))
			for indexSPK := range lft.SenderPublicKeys {
				signatures[indexSPK] = lft.Signatures[indexSPK].Signature
			}

			senderAccountIDs := make([]uint64, len(lft.SenderAccountIDs))
			for indexSA := range lft.SenderAccountIDs {
				senderAccountIDs[indexSA] = lft.SenderAccountIDs[indexSA].Uint64()
			}

			var bc *intMaxTypes.BlockContent
			bc, err = block_post_service.MakeNonRegistrationBlock(
				*lft.TxRoot,
				senderAccountIDs,
				lft.SenderPublicKeys,
				signatures,
			)
			if err != nil {
				return errors.Join(ErrMakeRegistrationBlockFail, err)
			}

			senders := make([]intMaxTypes.ColumnSender, 0)
			for i := range bc.Senders {
				senders = append(senders, intMaxTypes.ColumnSender{
					PublicKey: bc.Senders[i].PublicKey.ToAddress().String(),
					AccountID: bc.Senders[i].AccountID,
					IsSigned:  bc.Senders[i].IsSigned,
				})
			}

			var block *mDBApp.Block
			block, err = q.CreateBlock(
				mw.IntMaxPublicKey,
				hexutils.BytesToHex(bc.TxTreeRoot.Marshal()),
				hexutils.BytesToHex(bc.AggregatedSignature.Marshal()),
				hexutils.BytesToHex(bc.AggregatedPublicKey.Marshal()),
				senders,
				mDBApp.ST_ACCOUNT_ID,
				bytesLfsTree,
			)
			if err != nil {
				return errors.Join(ErrCreateBlockFail, err)
			}

			return funcLFT(block, lft)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (w *worker) TrHash(trHash string) (*TransactionHashesWithSenderAndFile, error) {
	w.trHashes.Lock()
	defer w.trHashes.Unlock()

	info, ok := w.trHashes.Hashes[trHash]
	if !ok {
		return nil, ErrTransactionHashNotFound
	}

	return info, nil
}

func (w *worker) SignTxTreeByAvailableFile(
	signature string,
	sf *TransactionHashesWithSenderAndFile,
	leafIndex uint64,
) error {
	f, ok := w.files.FilesList[sf.File]
	if !ok {
		// transfersHash not found
		return ErrTxTreeByAvailableFileFail
	}

	switch {
	case
		w.files.CurrentFile.Name() == sf.File.Name(),
		atomic.LoadInt32(&f.TransactionsCounter) != 0,
		!f.Delivered:
		// transfersHash exists, tx tree not found
		return ErrTxTreeNotFound
	case
		w.files.CurrentFile.Name() != sf.File.Name() &&
			atomic.LoadInt32(&f.TransactionsCounter) == 0 &&
			f.Delivered &&
			f.Timestamp != nil && f.Timestamp.UTC().Add(
			w.cfg.Worker.TimeoutForSignaturesAvailableFiles,
		).UnixNano() <= time.Now().UTC().UnixNano():
		for {
			if !f.Processing {
				// transfersHash exists, tx tree exists, signature collection for tx tree completed
				return ErrTxTreeSignatureCollectionComplete
			}
		}
	case
		w.files.CurrentFile.Name() != sf.File.Name() &&
			atomic.LoadInt32(&f.TransactionsCounter) == 0 &&
			f.Delivered &&
			f.Timestamp != nil && f.Timestamp.UTC().Add(
			w.cfg.Worker.TimeoutForSignaturesAvailableFiles,
		).UnixNano() > time.Now().UTC().UnixNano():
		for {
			if !f.Processing {
				break
			}
		}
	}

	w.files.Lock()
	defer w.files.Unlock()

	tm := time.Now().UTC().UnixNano()
	if w.files.FilesList[sf.File].Hashes[sf.TxHash].AccountID != nil &&
		w.files.FilesList[sf.File].LeafsTreeAccounts != nil {
		if vKeys, okKeys := w.files.FilesList[sf.File].LeafsTreeAccounts.KeysOfSenderPublicKeys[sf.Sender]; okKeys {
			w.files.FilesList[sf.File].LeafsTreeAccounts.Signatures[vKeys] = &signaturesByLeafIndex{
				Sender:    sf.Sender,
				TxHash:    sf.TxHash,
				Signature: signature,
				LeafIndex: leafIndex,
				CreatedAt: tm,
			}
			w.files.FilesList[sf.File].LeafsTreeAccounts.SignaturesByLeafIndex[leafIndex] = &signaturesByLeafIndex{
				Sender:    sf.Sender,
				TxHash:    sf.TxHash,
				Signature: signature,
				LeafIndex: leafIndex,
				CreatedAt: tm,
			}
			w.files.FilesList[sf.File].LeafsTreeAccounts.SignaturesCounter++
		}
	}

	if w.files.FilesList[sf.File].Hashes[sf.TxHash].AccountID == nil &&
		w.files.FilesList[sf.File].LeafsTreePublicKeys != nil {
		if vKeys, okKeys := w.files.FilesList[sf.File].LeafsTreePublicKeys.KeysOfSenderPublicKeys[sf.Sender]; okKeys {
			w.files.FilesList[sf.File].LeafsTreePublicKeys.Signatures[vKeys] = &signaturesByLeafIndex{
				Sender:    sf.Sender,
				TxHash:    sf.TxHash,
				Signature: signature,
				LeafIndex: leafIndex,
				CreatedAt: tm,
			}
			w.files.FilesList[sf.File].LeafsTreePublicKeys.SignaturesByLeafIndex[leafIndex] = &signaturesByLeafIndex{
				Sender:    sf.Sender,
				TxHash:    sf.TxHash,
				Signature: signature,
				LeafIndex: leafIndex,
				CreatedAt: tm,
			}
			w.files.FilesList[sf.File].LeafsTreePublicKeys.SignaturesCounter++
		}
	}

	return nil
}

func (w *worker) kvStore(filename string) (*bolt.DB, error) {
	const F0600 os.FileMode = 0600
	db, err := bolt.Open(filename, F0600, nil)
	if err != nil {
		return nil, errors.Join(ErrOpenFileKvStoreFail, err)
	}

	return db, nil
}
