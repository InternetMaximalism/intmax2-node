package mnemonic_wallet

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/mnemonic_wallet/models"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	isUtils "github.com/prodadidb/go-validation/is/utils"
	"github.com/tyler-smith/go-bip39"
)

type mnemonicWallet struct{}

func New() MnemonicWallet {
	return &mnemonicWallet{}
}

func (mw *mnemonicWallet) WalletGenerator(
	mnemonicDerivationPath, password string,
) (w *models.Wallet, err error) {
	const (
		minusKey = "-"
		emptyKey = ""
	)

	var mnemonic string
	mnemonic, err = bip39.NewMnemonic([]byte(
		strings.ReplaceAll(uuid.New().String(), minusKey, emptyKey),
	))
	if err != nil {
		return nil, errors.Join(ErrNewMnemonicFail, err)
	}

	w, err = mw.WalletFromMnemonic(mnemonic, password, mnemonicDerivationPath)
	if err != nil {
		return nil, errors.Join(ErrWalletFromMnemonicFail, err)
	}

	return w, nil
}

func (mw *mnemonicWallet) WalletFromMnemonic(
	mnemonic, password, mnemonicDerivationPath string,
) (w *models.Wallet, err error) {
	if mnemonic == "" {
		return nil, ErrMnemonicRequired
	}

	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, ErrMnemonicInvalid
	}

	var seed []byte
	seed, err = bip39.NewSeedWithErrorChecking(mnemonic, password)
	if err != nil {
		return nil, errors.Join(ErrNewSeedFail, err)
	}

	var masterKey *hdkeychain.ExtendedKey
	masterKey, err = hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, errors.Join(ErrNewMasterFromSeedFail, err)
	}

	var hdPath accounts.DerivationPath
	hdPath, err = accounts.ParseDerivationPath(mnemonicDerivationPath)
	if err != nil {
		return nil, errors.Join(ErrParseDerivationPathFail, err)
	}

	key := masterKey
	for _, n := range hdPath {
		key, err = key.Derive(n)
		if err != nil {
			return nil, errors.Join(ErrHDPathDeriveFail, err)
		}
	}

	var privateKey *btcec.PrivateKey
	privateKey, err = key.ECPrivKey()
	if err != nil {
		return nil, errors.Join(ErrECPrivateKeyFail, err)
	}

	privateKeyECDSA := privateKey.ToECDSA()

	var privateKeyHex string
	privateKeyHex, err = mw.privateKeyToHex(privateKeyECDSA)
	if err != nil {
		return nil, errors.Join(ErrProcessingPrivateKeyToHexFail, err)
	}

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, ErrPublicKeyFail
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	var intMaxPK *intMaxAcc.PrivateKey
	intMaxPK, err = intMaxAcc.NewINTMAXAccountFromECDSAKey(privateKeyECDSA)
	if err != nil {
		return nil, errors.Join(ErrNewINTMAXAccountFromECDSAKeyFail, err)
	}

	intMaxPublicKey := intMaxPK.PublicKey
	intMaxWalletAddress := intMaxPublicKey.ToAddress()
	intMaxPrivateKeyHex := intMaxPK.String()

	w = &models.Wallet{
		WalletAddress:       &address,
		PrivateKey:          privateKeyHex,
		Mnemonic:            mnemonic,
		DerivationPath:      mnemonicDerivationPath,
		Password:            password,
		IntMaxPublicKey:     intMaxPublicKey.String(),
		IntMaxWalletAddress: intMaxWalletAddress.String(),
		IntMaxPrivateKey:    intMaxPrivateKeyHex,
		PK:                  privateKeyECDSA,
	}

	return w, nil
}

func (mw *mnemonicWallet) WalletFromPrivateKeyHex(
	privateKeyHex string,
) (w *models.Wallet, err error) {
	if !isUtils.IsHexadecimal(privateKeyHex) {
		return nil, ErrPrivateKeyHexInvalid
	}

	var pk *ecdsa.PrivateKey
	pk, err = mw.privateKeyFromHex(privateKeyHex)
	if err != nil {
		return nil, errors.Join(ErrDecodePrivateKeyFromHEXToECDSAFail, err)
	}

	publicKey := pk.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, ErrPublicKeyFail
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	var intMaxPK *intMaxAcc.PrivateKey
	intMaxPK, err = intMaxAcc.NewINTMAXAccountFromECDSAKey(pk)
	if err != nil {
		return nil, errors.Join(ErrNewINTMAXAccountFromECDSAKeyFail, err)
	}

	intMaxPublicKey := intMaxPK.PublicKey
	intMaxWalletAddress := intMaxPublicKey.ToAddress()
	intMaxPrivateKeyHex := intMaxPK.String()

	w = &models.Wallet{
		WalletAddress:       &address,
		PrivateKey:          privateKeyHex,
		IntMaxPublicKey:     intMaxPublicKey.String(),
		IntMaxWalletAddress: intMaxWalletAddress.String(),
		IntMaxPrivateKey:    intMaxPrivateKeyHex,
		PK:                  pk,
	}

	return w, nil
}

func (mw *mnemonicWallet) privateKeyToHex(pk *ecdsa.PrivateKey) (h string, err error) {
	const (
		emptyKey = ""
	)

	defer func() {
		if r := recover(); r != nil {
			if err != nil {
				err = errors.Join(ErrHexEncodeFail, err)
			} else {
				err = ErrHexEncodeFail
			}
		}
	}()

	if pk == nil {
		return emptyKey, ErrPrivateKeyToHexFail
	}

	derivedKey := crypto.FromECDSA(pk)
	return hexutil.Encode(derivedKey)[2:], err
}

func (mw *mnemonicWallet) privateKeyFromHex(h string) (pk *ecdsa.PrivateKey, err error) {
	var decKey []byte
	decKey, err = hex.DecodeString(h)
	if err != nil {
		return nil, errors.Join(ErrHexDecodeFail, err)
	}

	pk, err = crypto.ToECDSA(decKey)
	if err != nil {
		return nil, errors.Join(ErrConvertPrivateKeyToECDSAFail, err)
	}

	return pk, nil
}
