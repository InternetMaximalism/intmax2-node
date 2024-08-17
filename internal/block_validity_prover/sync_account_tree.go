package block_validity_prover

import (
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/block_post_service"
	intMaxTree "intmax2-node/internal/tree"
	"math/big"
	"time"

	"github.com/holiman/uint256"
)

func (ai *MockBlockBuilder) RegisterPublicKey(pk *intMaxAcc.PublicKey, lastSentBlockNumber uint32) (accountID uint64, err error) {
	publicKey := pk.BigInt()
	proof, _, err := ai.AccountTree.ProveMembership(publicKey)
	if err != nil {
		var ErrProveMembershipFail = errors.New("failed to prove membership")
		return 0, errors.Join(ErrProveMembershipFail, err)
	}

	_, ok := ai.AccountTree.GetAccountID(publicKey)
	if ok {
		_, err = ai.AccountTree.Update(publicKey, uint64(0))
		if err != nil {
			var ErrUpdateAccountFail = errors.New("failed to update account")
			return 0, errors.Join(ErrUpdateAccountFail, err)
		}

		return uint64(proof.LeafIndex), nil
	}

	var insertionProof *intMaxTree.IndexedInsertionProof
	insertionProof, err = ai.AccountTree.Insert(publicKey, uint64(0))
	if err != nil {
		return 0, errors.Join(ErrCreateAccountFail, err)
	}

	return uint64(insertionProof.Index), nil
}

func (ai *MockBlockBuilder) PublicKeyByAccountID(accountID uint64) (pk *intMaxAcc.PublicKey, err error) {
	var accID uint256.Int
	accID.SetUint64(accountID)

	acc := ai.AccountTree.GetLeaf(accountID)

	pk, err = new(intMaxAcc.PublicKey).SetBigInt(acc.Key)
	if err != nil {
		return nil, errors.Join(ErrDecodeHexToPublicKeyFail, err)
	}

	return pk, nil
}

func (ai *MockBlockBuilder) AccountBySenderAddress(_ string) (*uint256.Int, error) {
	return nil, fmt.Errorf("AccountBySenderAddress not implemented")
}

func (p *blockValidityProver) SyncBlockTree(bps BlockSynchronizer) (err error) {
	blockNumber := p.blockBuilder.LastSeenBlockPostedEventBlockNumber

	var (
		events []*bindings.RollupBlockPosted
		nextBN *big.Int
	)

	events, nextBN, err = bps.FetchNewPostedBlocks(blockNumber)
	if err != nil {
		return errors.Join(ErrFetchNewPostedBlocksFail, err)
	}

	if len(events) == 0 {
		return nil
	}

	for key := range events {
		select {
		case <-p.ctx.Done():
			p.log.Warnf("Received cancel signal from context, stopping...")
			return p.ctx.Err()
		default:
			var blN uint256.Int
			_ = blN.SetFromBig(new(big.Int).SetUint64(events[key].Raw.BlockNumber))

			time.Sleep(1 * time.Second)

			var cd []byte
			cd, err = bps.FetchScrollCalldataByHash(events[key].Raw.TxHash)
			if err != nil {
				return errors.Join(ErrFetchScrollCalldataByHashFail, err)
			}

			postedBlock := block_post_service.NewPostedBlock(
				events[key].PrevBlockHash,
				events[key].DepositTreeRoot,
				uint32(events[key].BlockNumber.Uint64()),
				events[key].SignatureHash,
			)

			intMaxBlockNumber := events[key].BlockNumber

			blockHashLeaf := intMaxTree.NewBlockHashLeaf(postedBlock.Hash())
			blockTreeRoot, err := p.blockBuilder.BlockTree.AddLeaf(uint32(intMaxBlockNumber.Uint64()), blockHashLeaf)
			if err != nil {
				var ErrAddLeafFail = errors.New("failed to add leaf")
				return errors.Join(ErrAddLeafFail, err)
			}
			p.log.Debugf("Block tree root: %s\n", blockTreeRoot.String())

			_, err = FetchIntMaxBlockContentByCalldata(cd, postedBlock, p.blockBuilder)
			if err != nil {
				err = errors.Join(ErrFetchIntMaxBlockContentByCalldataFail, err)
				switch {
				case errors.Is(err, ErrUnknownAccountID):
					const msg = "block %q is ErrUnknownAccountID"
					p.log.WithError(err).Errorf(msg, intMaxBlockNumber.String())
				case errors.Is(err, ErrCannotDecodeAddress):
					const msg = "block %q is ErrCannotDecodeAddress"
					p.log.WithError(err).Errorf(msg, intMaxBlockNumber.String())
				default:
					const msg = "block %q processing error occurred"
					p.log.WithError(err).Errorf(msg, intMaxBlockNumber.String())
				}

				const msg = "processing of block %q error occurred"
				p.log.Debugf(msg, intMaxBlockNumber.String())

				continue
			}

			const msg = "block %q is valid (Scroll block number: %s)"
			p.log.Debugf(msg, intMaxBlockNumber.String(), blN.String())
		}
	}

	p.blockBuilder.LastSeenBlockPostedEventBlockNumber = nextBN.Uint64()

	return nil
}
