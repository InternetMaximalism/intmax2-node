package types

import (
	"encoding/binary"
	"encoding/json"
	"os"

	"github.com/iden3/go-iden3-crypto/ffg"
)

type Plonky2Proof struct {
	Proof        []byte        `json:"proof"`
	PublicInputs []ffg.Element `json:"public_inputs"`
}

func (p *Plonky2Proof) MarshalJSON() ([]byte, error) {
	publicInputs := make([]string, len(p.PublicInputs))
	for i, v := range p.PublicInputs {
		publicInputs[i] = v.String()
	}

	return json.Marshal(struct {
		Proof        []byte
		PublicInputs []string
	}{
		Proof:        p.Proof,
		PublicInputs: publicInputs,
	})
}

func (p *Plonky2Proof) UnmarshalJSON(data []byte) error {
	var aux struct {
		Proof        []byte
		PublicInputs []string
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	p.Proof = aux.Proof
	p.PublicInputs = make([]ffg.Element, len(aux.PublicInputs))
	for i, v := range aux.PublicInputs {
		p.PublicInputs[i].SetString(v)
	}

	return nil
}

func MakeSamplePlonky2Proof() (*Plonky2Proof, error) {
	proofBin, err := os.ReadFile("../../pkg/data/balance_proof.bin")
	if err != nil {
		return nil, err
	}

	publicInputsBin, err := os.ReadFile("../../pkg/data/balance_proof_public_inputs.bin")
	if err != nil {
		return nil, err
	}

	const numUint64Bytes = 8
	publicInputs := make([]ffg.Element, len(publicInputsBin)/numUint64Bytes)
	for i := 0; i < len(publicInputsBin)/numUint64Bytes; i += 1 {
		bytes := publicInputsBin[numUint64Bytes*i : numUint64Bytes*(i+1)]
		v := binary.LittleEndian.Uint64(bytes)
		publicInputs[i].SetUint64(v)
	}

	return &Plonky2Proof{
		Proof:        proofBin,
		PublicInputs: publicInputs,
	}, nil
}
