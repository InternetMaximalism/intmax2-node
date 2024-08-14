package handlers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"

	verifierCircuit "example.com/m/circuit"
	"example.com/m/context"
	"github.com/consensys/gnark-crypto/ecc"
	plonk_bn254 "github.com/consensys/gnark/backend/plonk/bn254"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/google/uuid"
	"github.com/qope/gnark-plonky2-verifier/types"
	"github.com/qope/gnark-plonky2-verifier/variables"
)

type ProveResult struct {
	PublicInputs []string `json:"publicInputs"`
	Proof string `json:"proof"`
}

type Status struct {
	Status string `json:"status"`
	Result ProveResult `json:"result"`
}

var (
    status = make(map[string]Status)
    mu      sync.Mutex
)

type CircuitData context.CircuitData


func extractPublicInputs(witness witness.Witness) ([]*big.Int, error) {
	public, err := witness.Public()
	if err != nil {
		return nil, err
	}
	_publicBytes, _ := public.MarshalBinary()
	publicBytes := _publicBytes[12:] 
	const chunkSize = 32 
	bigInts := make([]*big.Int, len(publicBytes)/chunkSize)
	for i := 0; i < len(publicBytes)/chunkSize; i += 1 {
		chunk := publicBytes[i*chunkSize : (i+1)*chunkSize]
		bigInt := new(big.Int).SetBytes(chunk)
		bigInts[i] = bigInt
	}
	return bigInts, nil
}


func (ctx *CircuitData) prove(jobId string, proofRaw types.ProofWithPublicInputsRaw) error{
	proofWithPis := variables.DeserializeProofWithPublicInputs(proofRaw)
	assignment := verifierCircuit.VerifierCircuit{
		Proof:                   proofWithPis.Proof,
		PublicInputs:            proofWithPis.PublicInputs,
		VerifierOnlyCircuitData: ctx.VerifierOnlyCircuitData,
	}
	witness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	if err != nil {
		mu.Lock()
		status[jobId] = Status{ Status:"error", Result:ProveResult{}}
		mu.Unlock()
		return err
	}
	proof, err := plonk_bn254.Prove(&ctx.Ccs, &ctx.Pk, witness)
	if err != nil {
		mu.Lock()
		status[jobId] = Status{ Status:"error", Result:ProveResult{}}
		mu.Unlock()
		return err
	}
	proofHex := hex.EncodeToString(proof.MarshalSolidity())
	publicInputs, err := extractPublicInputs(witness)
	if err != nil {
		mu.Lock()
		status[jobId] = Status{ Status:"error", Result:ProveResult{}}
		mu.Unlock()
		return err
	}
	publicInputsStr := make([]string, len(publicInputs))
	for i, bi := range publicInputs {
		publicInputsStr[i] = bi.String()
	}
	response := ProveResult{
		PublicInputs: publicInputsStr,
		Proof: proofHex,
	}
	mu.Lock()
	status[jobId] = Status{Result:response, Status:"done"}
	mu.Unlock()
	fmt.Println("Prove done. jobId", jobId)
	return nil
}

func (ctx *CircuitData) StartProof(w http.ResponseWriter, r *http.Request) {
	_jobId, err := uuid.NewRandom()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jobId := _jobId.String()
	var input types.ProofWithPublicInputsRaw
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mu.Lock()
	status[jobId] = Status{Status:"in progress"}
	mu.Unlock()
	go ctx.prove(jobId, input)
	json.NewEncoder(w).Encode(map[string]string{"jobId": jobId})
}

func (ctx *CircuitData) GetProof(w http.ResponseWriter, r *http.Request) {
	jobId := r.URL.Query().Get("jobId")
	_, err := uuid.Parse(jobId)
    if err != nil {
        http.Error(w, "Invalid JobId", http.StatusBadRequest)
        return
    }
	mu.Lock()
	defer mu.Unlock()
	s, ok := status[jobId]
	if !ok {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(s)
}
