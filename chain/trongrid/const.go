package trongrid

import (
	"github.com/joshuayildiz/wallet/chain"
)

// This value is basically precomputed keccak256('Transfer(address,address,uint256)')
// You can verify this by yourself if you want.
const encodedTransferEvent = "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

// Values below are taken from https://tether.to/en/supported-protocols/

func usdtContractAddr(net chain.Network) string {
	switch net {
	case chain.Mainnet:
		return "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	case chain.Testnet:
		return "TG3XXyExBkPp9nzdajDZsozEu4BkaSJozs"
	}
	return "unreachable"
}

func encodedUSDTContractAddr(net chain.Network) string {
	switch net {
	case chain.Mainnet:
		return "a614f803b6fd780986a42c78ec9c7f77e6ded13c"
	case chain.Testnet:
		return "42a1e39aefa49290f2b3f9ed688d7cecf86cd6e0"
	}
	return "unreachable"
}
