package block_validity_prover_server_test

import (
	"context"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	blockValidityProverBalanceUpdateWitness "intmax2-node/internal/use_cases/block_validity_prover_balance_update_witness"
	"intmax2-node/internal/use_cases/mocks"
	"intmax2-node/pkg/logger"
	ucBlockValidityProverBalanceUpdateWitness "intmax2-node/pkg/use_cases/block_validity_prover_balance_update_witness"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dimiro1/health"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"
)

func TestHandlerBalanceUpdateWitness(t *testing.T) {
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

	uc := mocks.NewMockUseCaseBlockValidityProverBalanceUpdateWitness(ctrl)

	randPH1, err := goldenposeidon.NewPoseidonHashOut().SetRandom()
	assert.NoError(t, err)
	randPH2, err := goldenposeidon.NewPoseidonHashOut().SetRandom()
	assert.NoError(t, err)
	randPH3, err := goldenposeidon.NewPoseidonHashOut().SetRandom()
	assert.NoError(t, err)
	siblings := []*goldenposeidon.PoseidonHashOut{
		randPH1,
		randPH2,
		randPH3,
	}

	accountMembershipProof := intMaxTree.IndexedMembershipProof{
		IsIncluded: true,
		LeafProof: intMaxTree.IndexedMerkleProof{
			Siblings: siblings,
		},
		LeafIndex: 10,
		Leaf: intMaxTree.IndexedMerkleLeaf{
			Key:       new(big.Int).SetUint64(123),
			Value:     11,
			NextIndex: 12,
			NextKey:   new(big.Int).SetUint64(321),
		},
	}

	accountMembershipProofWithEmptyLeafProof := intMaxTree.IndexedMembershipProof{
		IsIncluded: true,
		LeafProof:  intMaxTree.IndexedMerkleProof{},
		LeafIndex:  10,
		Leaf: intMaxTree.IndexedMerkleLeaf{
			Key:       new(big.Int).SetUint64(123),
			Value:     11,
			NextIndex: 12,
			NextKey:   new(big.Int).SetUint64(321),
		},
	}

	cases := []struct {
		desc       string
		body       string
		info       *blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness
		prepare    func(version *blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness)
		success    bool
		message    string
		wantStatus int
	}{
		{
			desc:       "currentBlockNumber: cannot be blank; targetBlockNumber: cannot be blank; user: cannot be blank.",
			message:    "currentBlockNumber: cannot be blank; targetBlockNumber: cannot be blank; user: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "currentBlockNumber: cannot be blank; targetBlockNumber: cannot be blank; user: must be a valid value.",
			body:       `{"user":"fake"}`,
			message:    "currentBlockNumber: cannot be blank; targetBlockNumber: cannot be blank; user: must be a valid value.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "currentBlockNumber: cannot be blank; targetBlockNumber: cannot be blank; user: must be a valid value.",
			body:       `{"user":"0x0000000000000000000000000000000000000000000000000000000000000000"}`,
			message:    "currentBlockNumber: cannot be blank; targetBlockNumber: cannot be blank; user: must be a valid value.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "currentBlockNumber: cannot be blank; targetBlockNumber: cannot be blank.",
			body:       `{"user":"0x0000000000000000000000000000000000000000000000000000000000000001"}`,
			message:    "currentBlockNumber: cannot be blank; targetBlockNumber: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "currentBlockNumber: must not be less than the targetBlockNumber value; targetBlockNumber: must not be more than the currentBlockNumber value.",
			body:       `{"user":"0x0000000000000000000000000000000000000000000000000000000000000001","currentBlockNumber":1,"targetBlockNumber":2}`,
			message:    "currentBlockNumber: must not be less than the targetBlockNumber value; targetBlockNumber: must not be more than the currentBlockNumber value.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: "currentBlockNumber: must not be less than the targetBlockNumber value; targetBlockNumber: must not be more than the currentBlockNumber value.",
			body: `{"user":"0x0000000000000000000000000000000000000000000000000000000000000001","currentBlockNumber":1,"targetBlockNumber":1}`,
			prepare: func(_ *blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness) {
				cmd.EXPECT().BlockValidityProverBalanceUpdateWitness(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(nil, ucBlockValidityProverBalanceUpdateWitness.ErrCurrentBlockNumberLessThenTargetBlockNumber)
			},
			message:    "currentBlockNumber: must not be less than the targetBlockNumber value; targetBlockNumber: must not be more than the currentBlockNumber value.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: "currentBlockNumber: must be a valid value.",
			body: `{"user":"0x0000000000000000000000000000000000000000000000000000000000000001","currentBlockNumber":1,"targetBlockNumber":1}`,
			prepare: func(_ *blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness) {
				cmd.EXPECT().BlockValidityProverBalanceUpdateWitness(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(nil, ucBlockValidityProverBalanceUpdateWitness.ErrCurrentBlockNumberInvalid)
			},
			message:    "currentBlockNumber: must be a valid value.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: "targetBlockNumber: must be a valid value.",
			body: `{"user":"0x0000000000000000000000000000000000000000000000000000000000000001","currentBlockNumber":1,"targetBlockNumber":1}`,
			prepare: func(_ *blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness) {
				cmd.EXPECT().BlockValidityProverBalanceUpdateWitness(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(nil, ucBlockValidityProverBalanceUpdateWitness.ErrTargetBlockNumberInvalid)
			},
			message:    "targetBlockNumber: must be a valid value.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: "user: must be a valid value.",
			body: `{"user":"0x0000000000000000000000000000000000000000000000000000000000000001","currentBlockNumber":1,"targetBlockNumber":1}`,
			prepare: func(_ *blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness) {
				cmd.EXPECT().BlockValidityProverBalanceUpdateWitness(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(nil, ucBlockValidityProverBalanceUpdateWitness.ErrPublicKeyFromIntMaxAccFail)
			},
			message:    "user: must be a valid value.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: "Internal server error",
			body: `{"user":"0x0000000000000000000000000000000000000000000000000000000000000001","currentBlockNumber":1,"targetBlockNumber":1}`,
			prepare: func(_ *blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness) {
				cmd.EXPECT().BlockValidityProverBalanceUpdateWitness(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(nil, errors.New("fake"))
			},
			message:    "Internal server error",
			wantStatus: http.StatusInternalServerError,
		},
		{
			desc: "success: empty resp.Data.BlockMerkleProof; empty resp.Data.AccountMembershipProof",
			body: `{"user":"0x0000000000000000000000000000000000000000000000000000000000000001","currentBlockNumber":2,"targetBlockNumber":1}`,
			info: &blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness{
				IsPrevAccountTree:      false,
				ValidityProof:          uuid.New().String(),
				BlockMerkleProof:       intMaxTree.BlockHashMerkleProof{},
				AccountMembershipProof: nil,
			},
			prepare: func(resp *blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness) {
				cmd.EXPECT().BlockValidityProverBalanceUpdateWitness(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(resp, nil)
			},
			success:    true,
			wantStatus: http.StatusOK,
		},
		{
			desc: "success: empty resp.Data.AccountMembershipProof",
			body: `{"user":"0x0000000000000000000000000000000000000000000000000000000000000001","currentBlockNumber":2,"targetBlockNumber":1}`,
			info: &blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness{
				IsPrevAccountTree: false,
				ValidityProof:     uuid.New().String(),
				BlockMerkleProof: intMaxTree.BlockHashMerkleProof{
					Siblings: siblings,
				},
				AccountMembershipProof: nil,
			},
			prepare: func(resp *blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness) {
				cmd.EXPECT().BlockValidityProverBalanceUpdateWitness(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(resp, nil)
			},
			success:    true,
			wantStatus: http.StatusOK,
		},
		{
			desc: "success: empty resp.Data.BlockMerkleProof",
			body: `{"user":"0x0000000000000000000000000000000000000000000000000000000000000001","currentBlockNumber":2,"targetBlockNumber":1}`,
			info: &blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness{
				IsPrevAccountTree:      false,
				ValidityProof:          uuid.New().String(),
				BlockMerkleProof:       intMaxTree.BlockHashMerkleProof{},
				AccountMembershipProof: &accountMembershipProof,
			},
			prepare: func(resp *blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness) {
				cmd.EXPECT().BlockValidityProverBalanceUpdateWitness(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(resp, nil)
			},
			success:    true,
			wantStatus: http.StatusOK,
		},
		{
			desc: "success: empty resp.Data.AccountMembershipProof.LeafProof.Siblings",
			body: `{"user":"0x0000000000000000000000000000000000000000000000000000000000000001","currentBlockNumber":2,"targetBlockNumber":1}`,
			info: &blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness{
				IsPrevAccountTree: false,
				ValidityProof:     uuid.New().String(),
				BlockMerkleProof: intMaxTree.BlockHashMerkleProof{
					Siblings: siblings,
				},
				AccountMembershipProof: &accountMembershipProofWithEmptyLeafProof,
			},
			prepare: func(resp *blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness) {
				cmd.EXPECT().BlockValidityProverBalanceUpdateWitness(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(resp, nil)
			},
			success:    true,
			wantStatus: http.StatusOK,
		},
		{
			desc: "success",
			body: `{"user":"0x0000000000000000000000000000000000000000000000000000000000000001","currentBlockNumber":2,"targetBlockNumber":1}`,
			info: &blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness{
				IsPrevAccountTree: false,
				ValidityProof:     uuid.New().String(),
				BlockMerkleProof: intMaxTree.BlockHashMerkleProof{
					Siblings: siblings,
				},
				AccountMembershipProof: &accountMembershipProof,
			},
			prepare: func(resp *blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness) {
				cmd.EXPECT().BlockValidityProverBalanceUpdateWitness(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(resp, nil)
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

			var body io.Reader
			body = http.NoBody
			if bd := strings.TrimSpace(cases[i].body); bd != "" {
				t.Log("-------- input body", bd)
				body = strings.NewReader(bd)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "http://"+gwServer.Addr+"/v1/balance-update-witness", body)

			gwServer.Handler.ServeHTTP(w, r)

			if !assert.Equal(t, cases[i].wantStatus, w.Code) {
				t.Log(w.Body.String())
			}

			assert.Equal(t, cases[i].message, strings.TrimSpace(gjson.Get(w.Body.String(), "message").String()))
			assert.Equal(t, cases[i].success, gjson.Get(w.Body.String(), "success").Bool())
			if cases[i].info != nil {
				assert.Equal(t, cases[i].info.IsPrevAccountTree, gjson.Get(w.Body.String(), "data.isPrevAccountTree").Bool())
				assert.Equal(t, cases[i].info.ValidityProof, gjson.Get(w.Body.String(), "data.validityProof").String())

				arrBlockMerkleProof := gjson.Get(w.Body.String(), "data.blockMerkleProof").Array()
				assert.Equal(t, len(cases[i].info.BlockMerkleProof.Siblings), len(arrBlockMerkleProof))
				for key := range arrBlockMerkleProof {
					assert.Equal(t, arrBlockMerkleProof[key].String(), cases[i].info.BlockMerkleProof.Siblings[key].String())
				}

				if cases[i].info.AccountMembershipProof == nil {
					assert.Nil(t, gjson.Get(w.Body.String(), "data.accountMembershipProof").Value())
				} else {
					assert.Equal(t, cases[i].info.AccountMembershipProof.IsIncluded, gjson.Get(w.Body.String(), "data.accountMembershipProof.isIncluded").Bool())

					arrAccountMembershipProofLeafProof := gjson.Get(w.Body.String(), "data.accountMembershipProof.leafProof").Array()
					assert.Equal(t, len(cases[i].info.AccountMembershipProof.LeafProof.Siblings), len(arrAccountMembershipProofLeafProof))
					for key := range arrAccountMembershipProofLeafProof {
						assert.Equal(t, arrAccountMembershipProofLeafProof[key].String(), cases[i].info.AccountMembershipProof.LeafProof.Siblings[key].String())
					}

					assert.Equal(t, cases[i].info.AccountMembershipProof.Leaf.Key.String(), gjson.Get(w.Body.String(), "data.accountMembershipProof.leaf.key").String())
					assert.Equal(t, int64(cases[i].info.AccountMembershipProof.Leaf.Value), gjson.Get(w.Body.String(), "data.accountMembershipProof.leaf.value").Int())
					assert.Equal(t, cases[i].info.AccountMembershipProof.Leaf.NextKey.String(), gjson.Get(w.Body.String(), "data.accountMembershipProof.leaf.nextKey").String())
					assert.Equal(t, int64(cases[i].info.AccountMembershipProof.Leaf.NextIndex), gjson.Get(w.Body.String(), "data.accountMembershipProof.leaf.nextIndex").Int())
				}
			}
		})
	}
}
