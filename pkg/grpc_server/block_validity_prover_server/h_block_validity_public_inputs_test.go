package block_validity_prover_server_test

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/block_validity_prover"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"
	blockValidityProverBlockValidityPublicInputs "intmax2-node/internal/use_cases/block_validity_prover_block_validity_public_inputs"
	"intmax2-node/internal/use_cases/mocks"
	"intmax2-node/pkg/logger"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"
)

func TestHandlerBlockValidityPublicInputs(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := configs.New()
	log := logger.New(cfg.LOG.Level, cfg.LOG.TimeFormat, cfg.LOG.JSON, cfg.LOG.IsLogLine)

	dbApp := NewMockSQLDriverApp(ctrl)
	hc := health.NewHandler()
	sb := NewMockServiceBlockchain(ctrl)
	bvs := NewMockBlockValidityService(ctrl)

	const (
		path1 = "../../../"
		path2 = "./"
	)

	dir := path1
	if _, err := os.ReadFile(dir + cfg.APP.PEMPathCACert); err != nil {
		dir = path2
	}
	cfg.APP.PEMPathCACert = dir + cfg.APP.PEMPathCACert
	cfg.APP.PEMPathServCert = dir + cfg.APP.PEMPathServCert
	cfg.APP.PEMPathServKey = dir + cfg.APP.PEMPathServKey
	cfg.APP.PEMPAthCACertClient = dir + cfg.APP.PEMPAthCACertClient
	cfg.APP.PEMPathClientCert = dir + cfg.APP.PEMPathClientCert
	cfg.APP.PEMPathClientKey = dir + cfg.APP.PEMPathClientKey

	cmd := NewMockCommands(ctrl)

	grpcServerStop, gwServer := Start(cmd, ctx, cfg, log, dbApp, &hc, sb, bvs)
	defer grpcServerStop()

	uc := mocks.NewMockUseCaseBlockValidityProverBlockValidityPublicInputs(ctrl)

	cases := []struct {
		desc        string
		blockNumber string
		info        *blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs
		prepare     func(*blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs)
		success     bool
		message     string
		wantStatus  int
	}{
		{
			desc:       `type mismatch, parameter: block_number, error: strconv.ParseUint: parsing "": invalid syntax`,
			message:    `type mismatch, parameter: block_number, error: strconv.ParseUint: parsing "": invalid syntax`,
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:        `type mismatch, parameter: block_number, error: strconv.ParseUint: parsing "-1": invalid syntax`,
			blockNumber: "-1",
			message:     `type mismatch, parameter: block_number, error: strconv.ParseUint: parsing "-1": invalid syntax`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			desc:        `blockNumber: cannot be blank.`,
			blockNumber: "0",
			message:     `blockNumber: cannot be blank.`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			desc:        `UC error: fake`,
			blockNumber: "1",
			prepare: func(inputs *blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs) {
				cmd.EXPECT().BlockValidityProverBlockValidityPublicInputs(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(nil, errors.New("fake"))
			},
			wantStatus: http.StatusOK,
		},
		{
			desc:        `empty UC result`,
			blockNumber: "1",
			prepare: func(inputs *blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs) {
				cmd.EXPECT().BlockValidityProverBlockValidityPublicInputs(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			desc:        `empty UC.ValidityPublicInputs result`,
			blockNumber: "1",
			info: &blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs{
				ValidityPublicInputs: nil,
				Sender:               nil,
			},
			prepare: func(inputs *blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs) {
				cmd.EXPECT().BlockValidityProverBlockValidityPublicInputs(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(inputs, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			desc:        `empty UC.ValidityPublicInputs.PublicState result`,
			blockNumber: "1",
			info: &blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs{
				ValidityPublicInputs: &block_validity_prover.ValidityPublicInputs{
					PublicState:    nil,
					TxTreeRoot:     intMaxTypes.Bytes32{},
					SenderTreeRoot: intMaxGP.NewPoseidonHashOut(),
					IsValidBlock:   false,
				},
				Sender: nil,
			},
			prepare: func(inputs *blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs) {
				cmd.EXPECT().BlockValidityProverBlockValidityPublicInputs(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(inputs, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			desc:        `empty UC.ValidityPublicInputs.PublicState.BlockTreeRoot result`,
			blockNumber: "1",
			info: &blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs{
				ValidityPublicInputs: &block_validity_prover.ValidityPublicInputs{
					PublicState: &block_validity_prover.PublicState{
						BlockTreeRoot:       nil,
						PrevAccountTreeRoot: intMaxGP.NewPoseidonHashOut(),
						AccountTreeRoot:     intMaxGP.NewPoseidonHashOut(),
						DepositTreeRoot:     common.Hash{},
						BlockHash:           common.Hash{},
						BlockNumber:         0,
					},
					TxTreeRoot:     intMaxTypes.Bytes32{},
					SenderTreeRoot: intMaxGP.NewPoseidonHashOut(),
					IsValidBlock:   false,
				},
				Sender: nil,
			},
			prepare: func(inputs *blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs) {
				cmd.EXPECT().BlockValidityProverBlockValidityPublicInputs(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(inputs, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			desc:        `empty UC.ValidityPublicInputs.PublicState.PrevAccountTreeRoot result`,
			blockNumber: "1",
			info: &blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs{
				ValidityPublicInputs: &block_validity_prover.ValidityPublicInputs{
					PublicState: &block_validity_prover.PublicState{
						BlockTreeRoot:       intMaxGP.NewPoseidonHashOut(),
						PrevAccountTreeRoot: nil,
						AccountTreeRoot:     intMaxGP.NewPoseidonHashOut(),
						DepositTreeRoot:     common.Hash{},
						BlockHash:           common.Hash{},
						BlockNumber:         0,
					},
					TxTreeRoot:     intMaxTypes.Bytes32{},
					SenderTreeRoot: intMaxGP.NewPoseidonHashOut(),
					IsValidBlock:   false,
				},
				Sender: nil,
			},
			prepare: func(inputs *blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs) {
				cmd.EXPECT().BlockValidityProverBlockValidityPublicInputs(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(inputs, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			desc:        `empty UC.ValidityPublicInputs.PublicState.AccountTreeRoot result`,
			blockNumber: "1",
			info: &blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs{
				ValidityPublicInputs: &block_validity_prover.ValidityPublicInputs{
					PublicState: &block_validity_prover.PublicState{
						BlockTreeRoot:       intMaxGP.NewPoseidonHashOut(),
						PrevAccountTreeRoot: intMaxGP.NewPoseidonHashOut(),
						AccountTreeRoot:     nil,
						DepositTreeRoot:     common.Hash{},
						BlockHash:           common.Hash{},
						BlockNumber:         0,
					},
					TxTreeRoot:     intMaxTypes.Bytes32{},
					SenderTreeRoot: intMaxGP.NewPoseidonHashOut(),
					IsValidBlock:   false,
				},
				Sender: nil,
			},
			prepare: func(inputs *blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs) {
				cmd.EXPECT().BlockValidityProverBlockValidityPublicInputs(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(inputs, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			desc:        `empty UC.ValidityPublicInputs.SenderTreeRoot result`,
			blockNumber: "1",
			info: &blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs{
				ValidityPublicInputs: &block_validity_prover.ValidityPublicInputs{
					PublicState:    &block_validity_prover.PublicState{},
					TxTreeRoot:     intMaxTypes.Bytes32{},
					SenderTreeRoot: nil,
					IsValidBlock:   false,
				},
				Sender: nil,
			},
			prepare: func(inputs *blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs) {
				cmd.EXPECT().BlockValidityProverBlockValidityPublicInputs(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(inputs, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			desc:        `success`,
			blockNumber: "1",
			info: &blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs{
				ValidityPublicInputs: &block_validity_prover.ValidityPublicInputs{
					PublicState: &block_validity_prover.PublicState{
						BlockTreeRoot:       intMaxGP.NewPoseidonHashOut(),
						PrevAccountTreeRoot: intMaxGP.NewPoseidonHashOut(),
						AccountTreeRoot:     intMaxGP.NewPoseidonHashOut(),
						DepositTreeRoot:     common.Hash{},
						BlockHash:           common.Hash{},
						BlockNumber:         0,
					},
					TxTreeRoot:     intMaxTypes.Bytes32{},
					SenderTreeRoot: intMaxGP.NewPoseidonHashOut(),
					IsValidBlock:   true,
				},
				Sender: []block_validity_prover.SenderLeaf{
					{
						Sender:  new(big.Int).SetUint64(123),
						IsValid: true,
					},
				},
			},
			prepare: func(inputs *blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs) {
				cmd.EXPECT().BlockValidityProverBlockValidityPublicInputs(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(inputs, nil)
			},
			success:    true,
			wantStatus: http.StatusOK,
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare(cases[i].info)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("http://"+gwServer.Addr+"/v1/block-validity-public-inputs/%s", cases[i].blockNumber),
				http.NoBody,
			)

			gwServer.Handler.ServeHTTP(w, r)

			if !assert.Equal(t, cases[i].wantStatus, w.Code) {
				t.Log(w.Body.String())
			}

			assert.Equal(t, cases[i].message, strings.TrimSpace(gjson.Get(w.Body.String(), "message").String()))
			assert.Equal(t, cases[i].success, gjson.Get(w.Body.String(), "success").Bool())
		})
	}
}
