package trongrid

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"

	"github.com/btcsuite/btcutil/base58"
	"github.com/joshuayildiz/wallet/chain"
)

func decodeTransferAddr(value string) string {
	addrBytes, err := hex.DecodeString(value)
	if err != nil {
		panic(err)
	}

	first := sha256.Sum256(addrBytes)
	second := sha256.Sum256(first[:])
	checksum := second[:4]

	both := append(addrBytes, checksum...)
	encoded := base58.Encode(both)

	return encoded
}

func decodeTopicAddr(net chain.Network, value string) string {
	last40 := value[24:]
	addrBytes, err := hex.DecodeString(last40)
	if err != nil {
		panic(err) // todo: should we panic here? figure that out
	}

	var networkedBuf bytes.Buffer
	switch net {
	case chain.Mainnet:
		networkedBuf.WriteByte(0x41)
	case chain.Testnet:
		networkedBuf.WriteByte(0x41) // apparently shasta also uses Mainnet -_-
	}
	networkedBuf.Write(addrBytes)

	networked := networkedBuf.Bytes()

	first := sha256.Sum256(networked)
	second := sha256.Sum256(first[:])
	checksum := second[:4]

	both := append(networked, checksum...)
	encoded := base58.Encode(both)

	return encoded
}
