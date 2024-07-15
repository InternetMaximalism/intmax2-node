package accounts

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/crypto/ecies"
)

func EncryptWithAES(key, plaintext []byte) (iv, ciphertext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	ciphertext = make([]byte, aes.BlockSize+len(plaintext))
	iv = ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return iv, ciphertext, nil
}

func DecryptWithAES(key, iv, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(ciphertext))
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

func EncryptWithAEAD(key, plaintext []byte) (nonce, ciphertext []byte, err error) {
	var block cipher.Block
	block, err = aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	var aesgcm cipher.AEAD
	aesgcm, err = cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce = make([]byte, aesgcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext = aesgcm.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}

func DecryptWithAEAD(key, nonce, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// NOTICE: This function uses deprecated code and is likely to change its implementation later.
func EncryptECIES(r io.Reader, pk *PublicKey, message []byte) (ciphertext []byte, err error) {
	return ecies.Encrypt(r, pk.ConvertECIESKey(), message, nil, nil)
}

// NOTICE: This function uses deprecated code and is likely to change its implementation later.
func (sk *PrivateKey) DecryptECIES(ciphertext []byte) (message []byte, err error) {
	return sk.ConvertECIESKey().Decrypt(ciphertext, nil, nil)
}

// The BN254G1Curve implements the elliptic.Curve interface for the G1 curve of BN254.
type BN254G1Curve struct{}

// Params returns the parameters for the curve.
func (c *BN254G1Curve) Params() *elliptic.CurveParams {
	const (
		int3Key   = 3
		int254Key = 254
	)

	generator := new(bn254.G1Affine).ScalarMultiplicationBase(big.NewInt(1))
	P := fp.Modulus()
	N := fr.Modulus()
	B := big.NewInt(int3Key)
	Gx := generator.X.BigInt(new(big.Int))
	Gy := generator.Y.BigInt(new(big.Int))
	BitSize := int254Key
	Name := "BN254G1"
	return &elliptic.CurveParams{
		P:       P,
		N:       N,
		B:       B,
		Gx:      Gx,
		Gy:      Gy,
		BitSize: BitSize,
		Name:    Name,
	}
}

// IsOnCurve reports whether the given (x,y) lies on the curve.
func (c *BN254G1Curve) IsOnCurve(x, y *big.Int) bool {
	x1Fp := new(fp.Element).SetBigInt(x)
	y1Fp := new(fp.Element).SetBigInt(y)

	p1 := bn254.G1Affine{X: *x1Fp, Y: *y1Fp}

	return p1.IsOnCurve()
}

// Add returns the sum of (x1,y1) and (x2,y2).
func (c *BN254G1Curve) Add(x1, y1, x2, y2 *big.Int) (x, y *big.Int) {
	x1Fp := new(fp.Element).SetBigInt(x1)
	y1Fp := new(fp.Element).SetBigInt(y1)
	x2Fp := new(fp.Element).SetBigInt(x2)
	y2Fp := new(fp.Element).SetBigInt(y2)

	p1 := bn254.G1Affine{X: *x1Fp, Y: *y1Fp}
	p2 := bn254.G1Affine{X: *x2Fp, Y: *y2Fp}

	p3 := new(bn254.G1Affine).Add(&p1, &p2)

	return p3.X.BigInt(new(big.Int)), p3.Y.BigInt(new(big.Int))
}

// Double returns 2*(x,y).
func (c *BN254G1Curve) Double(x1, y1 *big.Int) (x, y *big.Int) {
	x1Fp := new(fp.Element).SetBigInt(x1)
	y1Fp := new(fp.Element).SetBigInt(y1)

	p1 := bn254.G1Affine{X: *x1Fp, Y: *y1Fp}

	p3 := new(bn254.G1Affine).Double(&p1)

	return p3.X.BigInt(new(big.Int)), p3.Y.BigInt(new(big.Int))
}

// ScalarMult returns k*(x,y) where k is an integer in big-endian form.
func (c *BN254G1Curve) ScalarMult(x1, y1 *big.Int, k []byte) (x, y *big.Int) {
	x1Fp := new(fp.Element).SetBigInt(x1)
	y1Fp := new(fp.Element).SetBigInt(y1)

	p1 := bn254.G1Affine{X: *x1Fp, Y: *y1Fp}

	kFp := new(big.Int).SetBytes(k)

	p3 := new(bn254.G1Affine).ScalarMultiplication(&p1, kFp)

	return p3.X.BigInt(new(big.Int)), p3.Y.BigInt(new(big.Int))
}

// ScalarBaseMult returns k*G, where G is the base point of the group
// and k is an integer in big-endian form.
//
// Deprecated: this is a low-level unsafe API. For ECDH, use the crypto/ecdh
// package. Most uses of ScalarBaseMult can be replaced by a call to the
// PrivateKey.PublicKey method in crypto/ecdh.
func (c *BN254G1Curve) ScalarBaseMult(k []byte) (x, y *big.Int) {
	kFp := new(big.Int).SetBytes(k)

	p3 := new(bn254.G1Affine).ScalarMultiplicationBase(kFp)

	return p3.X.BigInt(new(big.Int)), p3.Y.BigInt(new(big.Int))
}

func (c *BN254G1Curve) Marshal(x, y *big.Int) []byte {
	const (
		int0Key = 0
		int1Key = 1
		int4Key = 4
	)

	xFp := new(fp.Element).SetBigInt(x)
	yFp := new(fp.Element).SetBigInt(y)

	p := bn254.G1Affine{X: *xFp, Y: *yFp}

	b := p.Marshal()
	res := make([]byte, int1Key+len(b))
	res[int0Key] = int4Key
	copy(res[int1Key:], b)

	return res
}

func (c *BN254G1Curve) Unmarshal(data []byte) (x, y *big.Int) {
	const (
		int0Key = 0
		int1Key = 1
		int4Key = 4
	)

	// mUncompressed
	if data[int0Key] != int4Key {
		fmt.Printf("invalid compression flag")
		return nil, nil
	}

	buffer := make([]byte, len(data)-int1Key)
	copy(buffer, data[int1Key:])

	p1 := new(bn254.G1Affine)
	err := p1.Unmarshal(buffer)
	if err != nil {
		fmt.Printf("invalid unmarshal: %v", err)
		return nil, nil
	}

	return p1.X.BigInt(new(big.Int)), p1.Y.BigInt(new(big.Int))
}

func (sk *PrivateKey) ConvertECIESKey() *ecies.PrivateKey {
	prv := new(ecies.PrivateKey)
	prv.PublicKey.X = sk.Public().Pk.X.BigInt(new(big.Int))
	prv.PublicKey.Y = sk.Public().Pk.Y.BigInt(new(big.Int))
	prv.PublicKey.Curve = &BN254G1Curve{}
	prv.D = new(big.Int).Set(sk.sk)
	prv.PublicKey.Params = ecies.ECIES_AES128_SHA256

	return prv
}

func (pk *PublicKey) ConvertECIESKey() *ecies.PublicKey {
	pub := new(ecies.PublicKey)
	pub.X = pk.Pk.X.BigInt(new(big.Int))
	pub.Y = pk.Pk.Y.BigInt(new(big.Int))
	pub.Curve = &BN254G1Curve{}
	pub.Params = ecies.ECIES_AES128_SHA256

	return pub
}
