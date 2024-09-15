package types

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"

	"github.com/iden3/go-iden3-crypto/ffg"
)

type Plonky2Proof struct {
	Proof        []byte        `json:"proof"`
	PublicInputs []ffg.Element `json:"public_inputs"`
	proofBytes   []byte
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

func NewCompressedPlonky2ProofFromBytes(proof []byte) (*Plonky2Proof, error) {
	reader := bufio.NewReader(bytes.NewReader(proof))
	numPublicInputs := uint32(0)
	err := binary.Read(reader, binary.LittleEndian, &numPublicInputs)
	if err != nil {
		return nil, err
	}

	publicInputs := make([]ffg.Element, numPublicInputs)
	for i := 0; i < int(numPublicInputs); i++ {
		v := uint64(0)
		binary.Read(reader, binary.LittleEndian, &v)
		publicInputs[i].SetUint64(v)
	}

	proofBytes := []byte{}
	_, err = reader.Read(proofBytes)
	if err != nil {
		return nil, err
	}

	return &Plonky2Proof{
		PublicInputs: publicInputs,
		Proof:        proof,
		proofBytes:   proofBytes,
	}, nil
}

func (p *Plonky2Proof) Bytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, uint32(len(p.PublicInputs)))
	if err != nil {
		return nil, err
	}

	for _, v := range p.PublicInputs {
		err = binary.Write(buf, binary.LittleEndian, v.ToUint64Regular())
		if err != nil {
			return nil, err
		}
	}

	_, err = buf.Write(p.Proof)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (p *Plonky2Proof) Base64String() (string, error) {
	bytes, err := p.Bytes()
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bytes), nil
}

func DecodePublicInputs(reader *bufio.Reader, numPublicInputs uint32) ([]ffg.Element, error) {
	publicInputs := make([]ffg.Element, numPublicInputs)
	for i := 0; i < int(numPublicInputs); i++ {
		v := uint64(0)
		err := binary.Read(reader, binary.LittleEndian, &v)
		if err != nil {
			return nil, err
		}

		publicInputs[i].SetUint64(v)
	}

	return publicInputs, nil
}

func NewCompressedPlonky2ProofFromBase64String(proof string) (*Plonky2Proof, error) {
	decodedProof, err := base64.StdEncoding.DecodeString(proof)
	if err != nil {
		return nil, errors.New("failed to decode transaction")
	}

	reader := bufio.NewReader(bytes.NewReader(decodedProof))
	numPublicInputs := uint32(0)
	err = binary.Read(reader, binary.LittleEndian, &numPublicInputs)
	if err != nil {
		return nil, err
	}

	publicInputs, err := DecodePublicInputs(reader, numPublicInputs)
	if err != nil {
		return nil, err
	}

	proofBytes := []byte{}
	_, err = reader.Read(proofBytes)
	if err != nil {
		return nil, err
	}

	return &Plonky2Proof{
		PublicInputs: publicInputs,
		Proof:        proofBytes,
	}, nil
}

func (p *Plonky2Proof) PublicInputsBytes() []byte {
	buf := new(bytes.Buffer)
	for _, v := range p.PublicInputs {
		binary.Write(buf, binary.LittleEndian, v.ToUint64Regular())
	}

	return buf.Bytes()
}

func (p *Plonky2Proof) ProofBase64String() string {
	return base64.StdEncoding.EncodeToString(p.Proof)
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
