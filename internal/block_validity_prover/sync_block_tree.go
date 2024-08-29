package block_validity_prover

import (
	"encoding/hex"
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/block_post_service"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
	"time"

	"github.com/holiman/uint256"
)

func (ai *mockBlockBuilder) RegisterPublicKey(pk *intMaxAcc.PublicKey, lastSentBlockNumber uint32) (accountID uint64, err error) {
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

func (ai *mockBlockBuilder) PublicKeyByAccountID(accountID uint64) (pk *intMaxAcc.PublicKey, err error) {
	var accID uint256.Int
	accID.SetUint64(accountID)

	acc := ai.AccountTree.GetLeaf(accountID)

	pk, err = new(intMaxAcc.PublicKey).SetBigInt(acc.Key)
	if err != nil {
		return nil, errors.Join(ErrDecodeHexToPublicKeyFail, err)
	}

	return pk, nil
}

func (ai *mockBlockBuilder) AccountBySenderAddress(_ string) (*uint256.Int, error) {
	return nil, fmt.Errorf("AccountBySenderAddress not implemented")
}

func (p *blockValidityProver) BlockBuilder() SQLDriverApp {
	return p.blockBuilder
}

func (p *blockValidityProver) SyncBlockTree(bps BlockSynchronizer) (err error) {
	blockNumber, err := p.blockBuilder.LastSeenBlockPostedEventBlockNumber()
	if err != nil {
		return errors.Join(ErrLastSeenBlockPostedEventBlockNumberFail, err)
	}

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

	tickerEventWatcher := time.NewTicker(p.cfg.BlockValidityProver.TimeoutForEventWatcher)
	defer func() {
		if tickerEventWatcher != nil {
			tickerEventWatcher.Stop()
		}
	}()
	for key := range events {
		select {
		case <-p.ctx.Done():
			p.log.Warnf("Received cancel signal from context, stopping...")
			return p.ctx.Err()
		case <-tickerEventWatcher.C:
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

			// Update account tree
			var blockContent *intMaxTypes.BlockContent
			blockContent, err = FetchIntMaxBlockContentByCalldata(cd, postedBlock, p.blockBuilder)
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
			} else {
				const msg = "block %q is valid (Scroll block number: %s)"
				p.log.Debugf(msg, intMaxBlockNumber.String(), blN.String())
			}

			// TODO: Update block hash tree

			senders := make([]intMaxTypes.ColumnSender, len(blockContent.Senders))
			for i, sender := range blockContent.Senders {
				senders[i] = intMaxTypes.ColumnSender{
					PublicKey: hex.EncodeToString(sender.PublicKey.ToAddress().Bytes()),
					AccountID: sender.AccountID,
					IsSigned:  sender.IsSigned,
				}
			}

			err := setAuxInfo(
				p.blockBuilder,
				postedBlock,
				blockContent,
			)
			if err != nil {
				return errors.Join(ErrCreateBlockContentFail, err)
			}

			// TODO: Separate another worker

			blockWitness, err := p.blockBuilder.GenerateBlock(blockContent, postedBlock)
			if err != nil {
				panic(err)
			}

			validityWitness, err := generateValidityWitness(p.blockBuilder, blockWitness)
			if err != nil {
				panic(err)
			}

			validityProof, err := p.requestAndFetchBlockValidityProof(validityWitness)
			if err != nil {
				return errors.Join(ErrRequestAndFetchBlockValidityProofFail, err)
			}

			p.blockBuilder.SetValidityProof(validityWitness.BlockWitness.Block.BlockNumber, validityProof)
		}
	}

	p.blockBuilder.SetLastSeenBlockPostedEventBlockNumber(nextBN.Uint64())

	return nil
}

func (p *blockValidityProver) requestAndFetchBlockValidityProof(validityWitness *ValidityWitness) (validityProof string, err error) {
	blockHash := validityWitness.BlockWitness.Block.Hash()
	lastValidityProof, err := p.blockBuilder.LastValidityProof()
	if err != nil && !errors.Is(err, ErrNoLastValidityProof) {
		var ErrLastValidityProofFail = errors.New("last validity proof fail")
		return "", errors.Join(ErrLastValidityProofFail, err)
	}
	err = p.requestBlockValidityProof(blockHash, validityWitness, lastValidityProof)
	if err != nil {
		var ErrRequestBlockValidityProofFail = errors.New("request block validity proof fail")
		return "", errors.Join(ErrRequestBlockValidityProofFail, err)
	}

	tickerBlockValidityProof := time.NewTicker(p.cfg.BlockValidityProver.TimeoutForFetchingBlockValidityProof)
	defer func() {
		if tickerBlockValidityProof != nil {
			tickerBlockValidityProof.Stop()
		}
	}()
	for {
		select {
		case <-p.ctx.Done():
			return "", p.ctx.Err()
		case <-tickerBlockValidityProof.C:
			var validityProof string
			validityProof, err = p.fetchBlockValidityProof(blockHash)
			if err != nil {
				continue
			}

			return validityProof, nil
		}
	}
}
