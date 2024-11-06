package block_validity_prover_balance_update_witness

import (
	"context"
	"errors"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucBlockValidityProverBalanceUpdateWitness "intmax2-node/internal/use_cases/block_validity_prover_balance_update_witness"

	"go.opentelemetry.io/otel/attribute"
)

type uc struct {
	cfg *configs.Config
	log logger.Logger
	bvs BlockValidityService
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	bvs BlockValidityService,
) ucBlockValidityProverBalanceUpdateWitness.UseCaseBlockValidityProverBalanceUpdateWitness {
	return &uc{
		cfg: cfg,
		log: log,
		bvs: bvs,
	}
}

func (u *uc) Do(
	ctx context.Context,
	input *ucBlockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitnessInput,
) (*ucBlockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness, error) {
	const (
		hName                 = "UseCase BalanceUpdateWitness"
		userKey               = "user"
		currentBlockNumberKey = "current_block_number"
		targetBlockNumberKey  = "target_block_number"
		isPrevAccountTreeKey  = "is_prev_account_tree"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCBlockValidityProverBalanceUpdateWitnessInputEmpty)
		return nil, ErrUCBlockValidityProverBalanceUpdateWitnessInputEmpty
	}

	span.SetAttributes(
		attribute.String(userKey, input.User),
		attribute.Int64(currentBlockNumberKey, input.CurrentBlockNumber),
		attribute.Int64(targetBlockNumberKey, input.TargetBlockNumber),
		attribute.Bool(isPrevAccountTreeKey, input.IsPrevAccountTree),
	)

	user, err := intMaxAcc.NewAddressFromHex(input.User)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrNewAddressFromHexFail, err)
	}

	var pubKey *intMaxAcc.PublicKey
	pubKey, err = user.Public()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrPublicKeyFromIntMaxAccFail, err)
	}

	var updW *block_validity_prover.UpdateWitness
	updW, err = u.bvs.FetchUpdateWitness(
		pubKey,
		uint32(input.CurrentBlockNumber),
		uint32(input.TargetBlockNumber),
		input.IsPrevAccountTree,
	)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		switch {
		case errors.Is(err, block_validity_prover.ErrRootBlockNumberLessThenLeafBlockNumber):
			return nil, errors.Join(ErrCurrentBlockNumberLessThenTargetBlockNumber, err)
		case
			errors.Is(err, block_validity_prover.ErrCurrentBlockNumberNotFound),
			errors.Is(err, block_validity_prover.ErrBlockTreeProofFail) &&
				errors.Is(err, block_validity_prover.ErrRootBlockNumberNotFound):
			return nil, errors.Join(ErrCurrentBlockNumberInvalid, err)
		case
			errors.Is(err, block_validity_prover.ErrBlockTreeProofFail) &&
				errors.Is(err, block_validity_prover.ErrLeafBlockNumberNotFound):
			return nil, errors.Join(ErrTargetBlockNumberInvalid, err)
		default:
			return nil, errors.Join(ErrFetchUpdateWitnessFail, err)
		}
	}

	return &ucBlockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness{
		IsPrevAccountTree:      input.IsPrevAccountTree,
		ValidityProof:          updW.ValidityProof,
		BlockMerkleProof:       updW.BlockMerkleProof,
		AccountMembershipProof: updW.AccountMembershipProof,
	}, nil
}
