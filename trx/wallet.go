package trx

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	decred_ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/joshuayildiz/wallet/chain"
	"github.com/joshuayildiz/wallet/chain/trongrid"
	"golang.org/x/crypto/sha3"
)

type Wallet struct {
	privKey  *secp256k1.PrivateKey
	trongrid *trongrid.Client
}

func New(trongrid *trongrid.Client) (*Wallet, error) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("trx.New: %w", err)
	}

	self := Wallet{
		privKey:  privKey,
		trongrid: trongrid,
	}
	return &self, nil
}

func NewWithPrivKeyHex(trongrid *trongrid.Client, privKeyHex string) (*Wallet, error) {
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("privkeyhex is invalid hex")
	}

	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)
	self := Wallet{
		privKey:  privKey,
		trongrid: trongrid,
	}

	return &self, nil
}

func (r *Wallet) PrivKeyHex() string {
	return hex.EncodeToString(r.privKey.Serialize())
}

func (r *Wallet) Addr() string {
	pubKey := r.privKey.PubKey().SerializeUncompressed()

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(pubKey[1:])
	keccak256 := hasher.Sum(nil)

	last20 := keccak256[len(keccak256)-20:]

	var networkedBuf bytes.Buffer
	switch r.trongrid.Net {
	case chain.Mainnet:
		networkedBuf.WriteByte(0x41)
	case chain.Testnet:
		networkedBuf.WriteByte(0x41) // apparently shasta also uses Mainnet -_-
	}
	networkedBuf.Write(last20)

	networked := networkedBuf.Bytes()

	first := sha256.Sum256(networked)
	second := sha256.Sum256(first[:])
	checksum := second[:4]

	both := append(networked, checksum...)
	encoded := base58.Encode(both)

	return encoded
}

func (r *Wallet) Balance(ctx context.Context) (uint, error) {
	balance, err := r.trongrid.Balance(ctx, r.Addr())
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (r *Wallet) Send(ctx context.Context, to string, amt uint) (string, error) {
	tx, err := r.trongrid.CreateTx(ctx, r.Addr(), to, amt)
	if err != nil {
		return "", fmt.Errorf("creating transaction: %w", err)
	}

	rawDataBytes, err := hex.DecodeString(tx.RawDataHex)
	if err != nil {
		return "", fmt.Errorf("raw data hex is invalid: %w", err)
	}

	sig, err := r.sign(rawDataBytes)
	if err != nil {
		return "", fmt.Errorf("signing raw data: %w", err)
	}

	tx.Signature = append(tx.Signature, sig)

	hash, err := r.trongrid.Broadcast(ctx, *tx)
	if err != nil {
		return "", fmt.Errorf("broadcasting tx: %w", err)
	}

	return hash, nil
}

func (r *Wallet) sign(data []byte) (string, error) {
	hash := sha256.Sum256(data)

	sig, err := sign(hash[:], r.privKey.ToECDSA())
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(sig), nil
}

// Sign calculates an ECDSA signature.
//
// This function is susceptible to chosen plaintext attacks that can leak
// information about the private key that is used for signing. Callers must
// be aware that the given hash cannot be chosen by an adversary. Common
// solution is to hash any input before calculating the signature.
//
// The produced signature is in the [R || S || V] format where V is 0 or 1.
func sign(hash []byte, prv *ecdsa.PrivateKey) ([]byte, error) {
	const DigestLength = 32
	const RecoveryIDOffset = 64

	if len(hash) != DigestLength {
		return nil, fmt.Errorf("hash is required to be exactly %d bytes (%d)", DigestLength, len(hash))
	}
	// ecdsa.PrivateKey -> secp256k1.PrivateKey
	var priv secp256k1.PrivateKey
	if overflow := priv.Key.SetByteSlice(prv.D.Bytes()); overflow || priv.Key.IsZero() {
		return nil, errors.New("invalid private key")
	}
	defer priv.Zero()
	sig := decred_ecdsa.SignCompact(&priv, hash, false) // ref uncompressed pubkey
	// Convert to Ethereum signature format with 'recovery id' v at the end.
	v := sig[0] - 27
	copy(sig, sig[1:])
	sig[RecoveryIDOffset] = v
	return sig, nil
}
