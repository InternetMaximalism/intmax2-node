package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"io/fs"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/uuid"
	"github.com/holiman/uint256"
	bolt "go.etcd.io/bbolt"
)

const (
	bucket     = "transfers"
	int1024Key = 1024
)

type kvInfo struct {
	Ctx                 context.Context
	CtxCancel           context.CancelFunc
	KvDB                *bolt.DB
	TransactionsCounter int32
	Delivered           bool
	Processing          bool
	Timestamp           *time.Time
	Receiver            chan func() error
}

type fileInfo struct {
	sync.Mutex
	kvInfo
	UsersCounter map[string]string
	Hashes       map[string]string
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
		Hashes:       make(map[string]string),
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

func (w *worker) AvailableFiles() (list []*os.File) {
	for key := range w.files.FilesList {
		cond1 := w.files.CurrentFile.Name() != key.Name()
		cond2 := atomic.LoadInt32(&w.files.FilesList[key].TransactionsCounter) == 0
		if cond1 && cond2 && !w.files.FilesList[key].Processing {
			fmt.Println("AvailableFiles cond1 && cond2")
			if !w.files.FilesList[key].Delivered {
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

	return list
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

	var kv *bolt.DB
	kv, err = w.kvStore(sf.File.Name())
	if err != nil {
		err = errors.Join(ErrKVStoreFail, err)
		return nil, err
	}
	defer func() {
		_ = kv.Close()
	}()

	err = kv.View(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket([]byte(bucket))
		s := b.Get([]byte(sf.Sender))

		err = json.Unmarshal(s, &txTreeRoot)
		if err != nil {
			return errors.Join(ErrUnmarshalFail, err)
		}

		return nil
	})
	if err != nil {
		return nil, errors.Join(ErrViewBucketKVStoreFail, err)
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
			list := w.AvailableFiles()
			fmt.Printf("tickerSignaturesAvailableFiles list %v\n", list)
			for key := range list {
				// cond1 - all transactions are processed
				cond1 := w.files.FilesList[list[key]].Delivered
				// cond2 - signature collection for tx tree completed
				cond2 := w.files.FilesList[list[key]].Timestamp != nil &&
					w.files.FilesList[list[key]].Timestamp.UTC().Add(
						w.cfg.Worker.TimeoutForSignaturesAvailableFiles,
					).UnixNano() < time.Now().UTC().UnixNano()
				if cond1 && cond2 {
					if atomic.LoadInt32(&w.numWorkers) < w.maxWorkers {
						fmt.Println("tickerSignaturesAvailableFiles cond1 && cond2")
						// Change status to processing
						w.files.FilesList[list[key]].Processing = true
						atomic.AddInt32(&w.numWorkers, 1)
						go func(f *os.File) {
							err := w.postProcessing(ctx, f)
							if err != nil {
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

	// crcs, err := w.currentRootCountAndSiblingsFromRW(input)
	// if err != nil {
	// 	return errors.Join(ErrCurrentRootCountAndSiblingsFromRW, err)
	// }

	transfersHashBytes, err := hexutil.Decode(input.TransferHash)
	if err != nil {
		var ErrHexDecodeFail = errors.New("fail to decode transfersHash")
		return errors.Join(ErrHexDecodeFail, err)
	}
	transfersHash := new(intMaxGP.PoseidonHashOut)
	transfersHash.Unmarshal(transfersHashBytes)

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
	w.files.FilesList[current].Hashes[input.TransferHash] = input.TransferHash

	w.files.FilesList[current].Receiver <- func() error {
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
			s          []byte
			crcs       *CurrentRootCountAndSiblings
			txTreeRoot TxTree
		)
		s = b.Get([]byte(input.Sender))
		if s != nil {
			err = json.Unmarshal(s, &txTreeRoot)
			if err != nil {
				return errors.Join(ErrUnmarshalFail, err)
			}
		} else {
			txTreeRoot = TxTree{}
			txTreeRoot.Sender = input.Sender
		}

		txTreeRoot.LeafIndexes = make(map[string]uint64)

		zeroHash := new(intMaxTypes.PoseidonHashOut).SetZero()
		var txTree *intMaxTree.TxTree
		txTree, err = intMaxTree.NewTxTree(intMaxTree.TX_TREE_HEIGHT, []*intMaxTypes.Tx{}, zeroHash)
		if err != nil {
			return errors.Join(ErrNewTxTreeFail, err)
		}

		crcs, err = w.currentRootCountAndSiblingsFromRW(input)
		if err != nil {
			return errors.Join(ErrCurrentRootCountAndSiblingsFromRW, err)
		}

		var st SenderTransfers
		st.ReceiverWorker = input
		st.CurrentRootCountAndSiblings = crcs

		fmt.Printf("st.ReceiverWorker %v\n", st.ReceiverWorker)
		txTreeRoot.SenderTransfers = append(txTreeRoot.SenderTransfers, &st)

		// var txTreeLeafHash []*intMaxTree.PoseidonHashOuts
		for key := range txTreeRoot.SenderTransfers {
			// var currTx *intMaxTypes.Tx
			// currTx, err = intMaxTypes.NewTx(
			// 	&txTreeRoot.SenderTransfers[key].CurrentRootCountAndSiblings.TransferTreeRoot, // XXX
			// 	txTreeRoot.SenderTransfers[key].ReceiverWorker.Nonce,
			// )
			// if err != nil {
			// 	return errors.Join(ErrNewTxFail, err)
			// }

			transfersHashBytes, err := hexutil.Decode(input.TransferHash)
			if err != nil {
				var ErrHexDecodeFail = errors.New("fail to decode transfersHash")
				return errors.Join(ErrHexDecodeFail, err)
			}
			transfersHash := new(intMaxTypes.PoseidonHashOut)
			err = transfersHash.Unmarshal(transfersHashBytes)
			if err != nil {
				return errors.Join(ErrUnmarshalFail, err)
			}

			var currTx *intMaxTypes.Tx
			currTx, err = intMaxTypes.NewTx(transfersHash, input.Nonce)
			if err != nil {
				return errors.Join(ErrNewTxFail, err)
			}

			txTreeRoot.SenderTransfers[key].TxHash = currTx.Hash()
			fmt.Printf("currTx %+v\n", currTx)
			fmt.Printf("txTreeRoot.SenderTransfers[key].TxHash %v\n", currTx.Hash())

			// txTreeRoot.SenderTransfers[key].TxTreeLeafHash, err = txTree.AddLeaf(uint64(key), currTx)
			_, err = txTree.AddLeaf(uint64(key), currTx)
			if err != nil {
				return errors.Join(ErrAddLeafIntoTxTreeFail, err)
			}
			// txTreeLeafHash = append(txTreeLeafHash, txTreeRootHash)

			txHash := txTreeRoot.SenderTransfers[key].ReceiverWorker.TxHash.Hash().String()
			txTreeRoot.LeafIndexes[txHash] = uint64(key)
		}

		txRoot, _, _ := txTree.GetCurrentRootCountAndSiblings()
		txTreeRoot.TxTreeHash = &txRoot
		// txTreeRoot.TxTreeHash, err = txTree.BuildMerkleRoot(txTreeLeafHash)
		if err != nil {
			return errors.Join(ErrTxTreeBuildMerkleRootFail, err)
		}

		for key := range txTreeRoot.SenderTransfers {
			txHash := txTreeRoot.SenderTransfers[key].ReceiverWorker.TxHash.Hash().String()
			index, ok := txTreeRoot.LeafIndexes[txHash] // = uint64(key)
			if !ok {
				var ErrLeafIndexNotFound = fmt.Errorf("leaf index not found")
				return ErrLeafIndexNotFound
			}
			fmt.Printf("transfersHash %v, index %v, key %v\n", txHash, index, key)
			var cmp ComputeMerkleProof
			var root intMaxTypes.PoseidonHashOut
			cmp.Siblings, root, err = txTree.ComputeMerkleProof(index)
			if err != nil {
				err = errors.Join(ErrTxTreeComputeMerkleProofFail, err)
				return err
			}

			if root != txRoot {
				fmt.Printf("root %v, txRoot %v\n", root, txRoot)
				var ErrTxTreeRootNotEqual = errors.New("tx tree root not equal")
				return ErrTxTreeRootNotEqual
			}
			txTreeRoot.SenderTransfers[key].TxTreeRootHash = &txRoot
			txTreeRoot.SenderTransfers[key].TxTreeSiblings = cmp.Siblings
		}

		var bST []byte
		bST, err = json.Marshal(&txTreeRoot)
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

		atomic.AddInt32(&w.files.FilesList[current].TransactionsCounter, -1)

		return nil
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

	var kv *bolt.DB
	kv, err = w.kvStore(f.Name())
	if err != nil {
		err = errors.Join(ErrKVStoreFail, err)
		return err
	}
	defer func() {
		_ = kv.Close()
	}()

	err = kv.View(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			err = w.dbApp.Exec(ctx, nil, func(d interface{}, _ interface{}) (err error) {
				q := d.(SQLDriverApp)

				var txTreeRoot TxTree
				err = json.Unmarshal(v, &txTreeRoot)
				if err != nil {
					err = errors.Join(ErrUnmarshalFail, err)
					return err
				}

				var txHashes []string
				for key := range txTreeRoot.SenderTransfers {
					txHashes = append(txHashes, txTreeRoot.SenderTransfers[key].ReceiverWorker.TxHash.Hash().String())
				}
				defer func() {
					for keyT := range txHashes {
						delete(w.files.FilesList[f].Hashes, txHashes[keyT])
						w.trHashes.Cleaner <- func() {
							w.trHashes.Lock()
							defer w.trHashes.Unlock()
							delete(w.trHashes.Hashes, txHashes[keyT])
						}
					}
				}()

				fmt.Printf("signature %v\n", txTreeRoot.Signature)
				const emptySignature = ""
				if txTreeRoot.Signature == emptySignature {
					fmt.Println("emptySignature")
					return nil
				}

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

				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (w *worker) currentRootCountAndSiblingsFromRW(
	rw *ReceiverWorker,
) (*CurrentRootCountAndSiblings, error) {
	transferTree, err := intMaxTree.NewTransferTree(
		intMaxTree.TX_TREE_HEIGHT,
		rw.TransferData,
		intMaxGP.NewPoseidonHashOut(),
	)
	if err != nil {
		return nil, errors.Join(ErrNewTransferTreeFail, err)
	}

	transferRoot, count, siblings := transferTree.GetCurrentRootCountAndSiblings()

	return &CurrentRootCountAndSiblings{
		TransferTreeRoot: transferRoot,
		Count:            count,
		Siblings:         siblings,
	}, err
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

func (w *worker) kvStore(filename string) (*bolt.DB, error) {
	const F0600 os.FileMode = 0600
	db, err := bolt.Open(filename, F0600, nil)
	if err != nil {
		return nil, errors.Join(ErrOpenFileKvStoreFail, err)
	}

	return db, nil
}
