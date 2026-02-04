package trongrid

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/joshuayildiz/wallet/chain"
)

// ValidateAddr validates a TRON address format and checksum.
// Returns nil if valid, otherwise returns an error describing the issue.
func ValidateAddr(addr string) error {
	if len(addr) != 34 {
		return fmt.Errorf("invalid address length: got %d, want 34", len(addr))
	}

	if addr[0] != 'T' {
		return fmt.Errorf("invalid address prefix: got %q, want 'T'", addr[0])
	}

	decoded := base58.Decode(addr)
	if len(decoded) != 25 {
		return fmt.Errorf("invalid decoded address length: got %d, want 25", len(decoded))
	}

	// Verify checksum (last 4 bytes)
	payload := decoded[:21]
	checksum := decoded[21:]

	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	expectedChecksum := second[:4]

	if !bytes.Equal(checksum, expectedChecksum) {
		return fmt.Errorf("invalid address checksum")
	}

	return nil
}

func decodeTransferAddr(value string) (string, error) {
	addrBytes, err := hex.DecodeString(value)
	if err != nil {
		return "", fmt.Errorf("decoding transfer address: %w", err)
	}

	first := sha256.Sum256(addrBytes)
	second := sha256.Sum256(first[:])
	checksum := second[:4]

	both := append(addrBytes, checksum...)
	encoded := base58.Encode(both)

	return encoded, nil
}

func decodeTopicAddr(net chain.Network, value string) (string, error) {
	last40 := value[24:]
	addrBytes, err := hex.DecodeString(last40)
	if err != nil {
		return "", fmt.Errorf("decoding topic address: %w", err)
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

	return encoded, nil
}
