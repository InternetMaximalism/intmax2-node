package models

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Wallet struct {
	WalletAddress       *common.Address   `json:"wallet_address,omitempty"`
	PrivateKey          string            `json:"private_key,omitempty"`
	DerivationPath      string            `json:"derivation_path,omitempty"`
	Mnemonic            string            `json:"mnemonic,omitempty"`
	Password            string            `json:"password,omitempty"`
	IntMaxWalletAddress string            `json:"intmax_wallet_address"`
	IntMaxPublicKey     string            `json:"intmax_public_key"`
	IntMaxPrivateKey    string            `json:"intmax_private_key"`
	PK                  *ecdsa.PrivateKey `json:"-"`
}

func (w *Wallet) Marshal() ([]byte, error) {
	if w == nil {
		return json.Marshal(&Wallet{})
	}

	return json.Marshal(*w)
}

func (w *Wallet) Unmarshal(input []byte) (err error) {
	if w == nil || input == nil {
		return nil
	}

	err = json.Unmarshal(input, w)
	if err != nil {
		return err
	}

	return nil
}

func (w *Wallet) Pk() (pk *ecdsa.PrivateKey, err error) {
	pk, err = crypto.HexToECDSA(w.PrivateKey)
	if err != nil {
		return nil, errors.Join(ErrHexToECDSAFail, err)
	}

	return pk, nil
}
