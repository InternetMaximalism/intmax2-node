package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"io/fs"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
)

const (
	bucket     = "transfers"
	int1024Key = 1024
)

type LeafsTree struct {
	TxTree     *intMaxTree.TxTree
	TxRoot     *intMaxTree.PoseidonHashOut
	Count      uint64
	Siblings   []*intMaxTree.PoseidonHashOut
	Signatures []string
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
	LeafsTree           *LeafsTree
}

type leafsOfHash struct {
	Sender string
	Index  uint64
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
				w.files.FilesList[currentFile].Delivered = true
				_ = w.files.FilesList[currentFile].KvDB.Close()
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
	siblings, root, err = f.LeafsTree.TxTree.ComputeMerkleProof(f.Hashes[sf.TxHash].Index)
	if err != nil {
		return nil, errors.Join(ErrTxTreeComputeMerkleProofFail, err)
	}

	txTreeRoot = &TxTree{
		RootHash: &root,
		Siblings: siblings,
	}

	return txTreeRoot, err
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
	atomic.AddInt32(&w.files.FilesList[current].TransactionsCounter, 1)
	w.files.FilesList[current].UsersCounter[input.Sender] = input.Sender
	w.files.FilesList[current].Hashes[input.TxHash.Hash().String()] = nil

	w.files.FilesList[current].Receiver <- func() error {
		defer atomic.AddInt32(&w.files.FilesList[current].TransactionsCounter, -1)

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

	zeroHash := new(intMaxTypes.PoseidonHashOut).SetZero()
	var txTree *intMaxTree.TxTree
	txTree, err = intMaxTree.NewTxTree(intMaxTree.TX_TREE_HEIGHT, []*intMaxTypes.Tx{}, zeroHash)
	if err != nil {
		return errors.Join(ErrNewTxTreeFail, err)
	}

	err = w.files.FilesList[f].KvDB.View(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}

		c := b.Cursor()
		var number int
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var info SenderInfo
			err = json.Unmarshal(v, &info)
			if err != nil {
				err = errors.Join(ErrUnmarshalFail, err)
				return err
			}

			for key := range info.TxsList {
				cn := uint64(number)
				_, err = txTree.AddLeaf(cn, info.TxsList[key].TxHash)
				if err != nil {
					return errors.Join(ErrAddLeafIntoTxTreeFail, err)
				}
				w.files.FilesList[f].Hashes[info.TxsList[key].TxHash.Hash().String()] = &leafsOfHash{
					Sender: info.TxsList[key].Sender,
					Index:  cn,
				}
				number++
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	txRoot, count, sb := txTree.GetCurrentRootCountAndSiblings()

	w.files.FilesList[f].Lock()
	defer w.files.FilesList[f].Unlock()

	w.files.FilesList[f].LeafsTree = &LeafsTree{
		TxTree:   txTree,
		TxRoot:   &txRoot,
		Count:    count,
		Siblings: sb,
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
		w.files.FilesList[f].Processing = false
	}()

	for key := range w.files.FilesList[f].Hashes {
		err = w.dbApp.Exec(ctx, nil, func(d interface{}, _ interface{}) (err error) {
			q := d.(SQLDriverApp)
			_ = q

			var (
				siblings []*intMaxTree.PoseidonHashOut
				root     intMaxTree.PoseidonHashOut
			)
			siblings, root, err = w.files.FilesList[f].LeafsTree.
				TxTree.ComputeMerkleProof(w.files.FilesList[f].Hashes[key].Index)
			if err != nil {
				return errors.Join(ErrTxTreeComputeMerkleProofFail, err)
			}

			_ = siblings
			_ = root

			defer func() {
				delete(w.files.FilesList[f].Hashes, key)
				w.trHashes.Cleaner <- func() {
					w.trHashes.Lock()
					defer w.trHashes.Unlock()
					delete(w.trHashes.Hashes, key)
				}
			}()

			const emptySignature = 0
			if len(w.files.FilesList[f].LeafsTree.Signatures) == emptySignature {
				return nil
			}

			/**
			var sign *mDBApp.Signature
			sign, err = q.CreateSignature(txTreeRoot.Signature)
			if err != nil {
				err = errors.Join(ErrCreateSignatureFail, err)
				return err
			}

			var publicKey *intMaxAcc.PublicKey
			publicKey, err = intMaxAcc.NewPublicKeyFromAddressHex(txTreeRoot.Sender)
			if err != nil {
				err = errors.Join(ErrNewPublicKeyFromAddressHexFail, err)
				return err
			}

			var transaction *mDBApp.Transactions
			transaction, err = q.CreateTransaction(
				publicKey.String(), txTreeRoot.TxTreeHash.String(), sign.SignatureID,
			)
			if err != nil {
				err = errors.Join(ErrCreateTransactionFail, err)
				return err
			}

			for key := range txTreeRoot.SenderTransfers {
				_, ok := w.files.FilesList[f].Hashes[txTreeRoot.SenderTransfers[key].ReceiverWorker.TxHash.Hash().String()]
				if !ok {
					err = ErrTransactionHashNotFound
					return err
				}

				var txIndex uint256.Int
				_ = txIndex.SetUint64(
					txTreeRoot.LeafIndexes[txTreeRoot.SenderTransfers[key].ReceiverWorker.TxHash.Hash().String()],
				)

				var cmp ComputeMerkleProof
				cmp.Root = *txTreeRoot.SenderTransfers[key].TxTreeRootHash
				cmp.Siblings = txTreeRoot.SenderTransfers[key].TxTreeSiblings

				var bCMP []byte
				bCMP, err = json.Marshal(&cmp)
				if err != nil {
					err = errors.Join(ErrMarshalFail, err)
					return err
				}

				_, err = q.CreateTxMerkleProofs(
					publicKey.String(),
					txTreeRoot.SenderTransfers[key].TxHash.String(),
					transaction.TxID,
					&txIndex,
					bCMP,
				)
				if err != nil {
					err = errors.Join(ErrCreateTxMerkleProofsFail, err)
					return err
				}
			}
			*/

			return nil
		})
		if err != nil {
			return err
		}
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

func (w *worker) SignTxTreeByAvailableFile(signature string, sf *TransactionHashesWithSenderAndFile) error {
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

	// TODO: sign txHash in TxTree by AvailableFile

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
