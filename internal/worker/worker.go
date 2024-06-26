package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/tree"
	"io/fs"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/holiman/uint256"
	bolt "go.etcd.io/bbolt"
)

const (
	bucket     = "transfers"
	int1024Key = 1024
)

type kvInfo struct {
	Ctx              context.Context
	CtxCancel        context.CancelFunc
	KvDB             *bolt.DB
	TransfersCounter int32
	Processing       bool
	Receiver         chan func() error
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

type transferHashesList struct {
	sync.Mutex
	Hashes  map[string]string
	Cleaner chan func()
}

type worker struct {
	cfg        *configs.Config
	log        logger.Logger
	dbApp      SQLDriverApp
	files      *workerFileList
	trHashes   *transferHashesList
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
		trHashes: &transferHashesList{
			Hashes:  make(map[string]string),
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

	w.files.CurrentFile = currentFile

	go func() {
		for {
			select {
			case <-w.files.FilesList[currentFile].Ctx.Done():
				if w.files != nil &&
					w.files.FilesList[currentFile] != nil &&
					w.files.FilesList[currentFile].Receiver != nil {
					close(w.files.FilesList[currentFile].Receiver)
					_ = w.files.FilesList[currentFile].KvDB.Close()
					w.files.FilesList[currentFile].Receiver = nil
				}
				return
			case fn := <-w.files.FilesList[currentFile].Receiver:
				errTx := fn()
				if errTx != nil {
					fmt.Println(errTx)
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
		if w.files.CurrentFile.Name() != key.Name() &&
			atomic.LoadInt32(&w.files.FilesList[key].TransfersCounter) == 0 &&
			!w.files.FilesList[key].Processing {
			w.files.FilesList[key].CtxCancel()
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

func (w *worker) Start(ctx context.Context, ticker *time.Ticker) error {
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return nil
		case <-ticker.C:
			st, err := w.files.CurrentFile.Stat()
			if err != nil {
				return errors.Join(ErrStatCurrentFileFail, err)
			}

			if st.ModTime().UTC().Add(
				w.cfg.Worker.CurrentFileLifetime,
			).UnixNano()-time.Now().UTC().UnixNano() <= 0 ||
				len(w.files.FilesList[w.files.CurrentFile].UsersCounter) > w.cfg.Worker.MaxCounterOfUsers {
				err = w.newTempFile(w.files.CurrentDir)
				if err != nil {
					return errors.Join(ErrCreateNewTempFileFail, err)
				}
			}

			list := w.AvailableFiles()
			if len(list) > 0 {
				if atomic.LoadInt32(&w.numWorkers) < w.maxWorkers {
					w.files.FilesList[list[0]].Processing = true
					atomic.AddInt32(&w.numWorkers, 1)
					go func() {
						err = w.postProcessing(ctx, list[0])
						if err != nil {
							const msg = "failed to apply post processing"
							w.log.WithError(err).Errorf(msg)
						}
					}()
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

	w.trHashes.Lock()
	defer w.trHashes.Unlock()

	_, ok := w.trHashes.Hashes[input.TransferHash]
	if ok {
		return ErrReceiverWorkerDuplicate
	}

	w.trHashes.Hashes[input.TransferHash] = input.Sender

	err := w.registerReceiver(input)
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
	atomic.AddInt32(&w.files.FilesList[current].TransfersCounter, 1)
	w.files.FilesList[current].UsersCounter[input.Sender] = input.Sender
	w.files.FilesList[current].Hashes[input.TransferHash] = input.TransferHash

	var bi []byte
	bi, err = json.Marshal(input)
	if err != nil {
		return errors.Join(ErrMarshalFail, err)
	}

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

		err = b.Put([]byte(input.TransferHash), bi)
		if err != nil {
			return errors.Join(ErrPutBucketKVStoreFail, err)
		}

		err = tx.Commit()
		if err != nil {
			return errors.Join(ErrTxCommitKVStoreFail, err)
		}

		atomic.AddInt32(&w.files.FilesList[current].TransfersCounter, -1)

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
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			err = w.dbApp.Exec(ctx, nil, func(d interface{}, _ interface{}) (err error) {
				q := d.(SQLDriverApp)

				var rw ReceiverWorker
				err = json.Unmarshal(v, &rw)
				if err != nil {
					err = errors.Join(ErrUnmarshalFail, err)
					return err
				}

				_, ok := w.files.FilesList[f].Hashes[rw.TransferHash]
				if !ok {
					return nil
				}

				var srcs *CurrentRootCountAndSiblings
				srcs, err = w.CurrentRootCountAndSiblingsFromRW(&rw)
				if err != nil {
					err = errors.Join(ErrCurrentRootCountAndSiblingsFromRW, err)
					return err
				}

				var bytesSRCS []byte
				bytesSRCS, err = json.Marshal(&srcs)
				if err != nil {
					err = errors.Join(ErrMarshalFail, err)
					return err
				}

				var txIndex uint256.Int
				_ = txIndex.SetUint64(uint64(len(rw.TransferData)))
				_, err = q.CreateTxMerkleProofs(rw.Sender, rw.TransferHash, &txIndex, bytesSRCS)
				if err != nil {
					err = errors.Join(ErrCreateTxMerkleProofsFail, err)
					return err
				}

				delete(w.files.FilesList[f].Hashes, rw.TransferHash)

				w.trHashes.Cleaner <- func() {
					w.trHashes.Lock()
					defer w.trHashes.Unlock()
					delete(w.trHashes.Hashes, rw.TransferHash)
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

func (w *worker) CurrentRootCountAndSiblingsFromRW(
	rw *ReceiverWorker,
) (*CurrentRootCountAndSiblings, error) {
	transferTree, err := tree.NewTransferTree(
		uint8(len(rw.TransferData)),
		rw.TransferData,
		intMaxGP.NewPoseidonHashOut(),
	)
	if err != nil {
		return nil, errors.Join(ErrNewTransferTreeFail, err)
	}

	transferRoot, count, siblings := transferTree.GetCurrentRootCountAndSiblings()

	return &CurrentRootCountAndSiblings{
		TransferRoot: transferRoot,
		Count:        count,
		Siblings:     siblings,
	}, err
}

func (w *worker) kvStore(filename string) (*bolt.DB, error) {
	const F0600 os.FileMode = 0600
	db, err := bolt.Open(filename, F0600, nil)
	if err != nil {
		return nil, errors.Join(ErrOpenFileKvStoreFail, err)
	}

	return db, nil
}
