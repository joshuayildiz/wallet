package trongrid

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/joshuayildiz/wallet/chain"
)

type Client struct {
	Net    chain.Network
	apikey string
	client *http.Client
}

func New(net chain.Network, apikey string) *Client {
	retryableClient := retryablehttp.NewClient()
	retryableClient.RetryMax = 3
	retryableClient.Logger = nil
	return &Client{
		Net:    net,
		apikey: apikey,
		client: retryableClient.StandardClient(),
	}
}

func (r *Client) Balance(ctx context.Context, addr string) (uint, error) {
	body := map[string]any{
		"address": addr,
		"visible": true,
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, r.url("/wallet/getaccount"), bytes.NewBuffer(bodyBytes))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("TRON-PRO-API-KEY", r.apikey)

	resp, err := r.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fetching balance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("fetching balance: %s", resp.Status)
	}

	var data struct {
		Balance uint `json:"balance"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return 0, fmt.Errorf("decoding balance: %w", err)
	}

	return data.Balance, nil
}

func (r *Client) Now(ctx context.Context) (*Block, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet, r.url("/walletsolidity/getnowblock"),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("TRON-PRO-API-KEY", r.apikey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching now block: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching now block: %s", resp.Status)
	}

	var data Block
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("decoding now block: %w", err)
	}

	return &data, nil
}

func (r *Client) BlockByNum(ctx context.Context, num uint) (*Block, error) {
	body := map[string]any{"num": num}
	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost, r.url("/walletsolidity/getblockbynum"),
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("TRON-PRO-API-KEY", r.apikey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching block %d: %w", num, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching block %d: %s", num, resp.Status)
	}

	var data Block
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("decoding block %d: %w", num, err)
	}

	return &data, nil
}

func (r *Client) TxInfoByBlockNum(ctx context.Context, num uint) ([]TxInfo, error) {
	body := map[string]any{"num": num}
	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost, r.url("/walletsolidity/gettransactioninfobyblocknum"),
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("TRON-PRO-API-KEY", r.apikey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching tx info by block num %d: %w", num, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching tx info by block num %d: %s", num, resp.Status)
	}

	var data []TxInfo
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("decoding tx info by block num %d: %w", num, err)
	}

	return data, nil
}

func (r *Client) TxInfoByID(ctx context.Context, id string) (*TxInfo, error) {
	body := map[string]any{"value": id}
	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost, r.url("/walletsolidity/gettransactioninfobyid"),
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("TRON-PRO-API-KEY", r.apikey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching tx by id %s: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching tx by id %s: %s", id, resp.Status)
	}

	var data TxInfo
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("decoding tx by id %s: %w", id, err)
	}

	return &data, nil
}

func (r *Client) CreateTx(ctx context.Context, from, to string, amt uint) (*Tx, error) {
	body := map[string]any{
		"owner_address": from,
		"to_address":    to,
		"amount":        amt,
		"visible":       true,
	}
	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost, r.url("/wallet/createtransaction"),
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("TRON-PRO-API-KEY", r.apikey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("creating tx: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("creating tx: %s", resp.Status)
	}

	var data Tx
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("decoding created tx: %w", err)
	}

	return &data, nil
}

func (r *Client) Broadcast(ctx context.Context, tx Tx) (string, error) {
	bodyBytes, _ := json.Marshal(tx)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost, r.url("/wallet/broadcasttransaction"),
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("TRON-PRO-API-KEY", r.apikey)

	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("broadcasting tx: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("broadcasting tx: %s", resp.Status)
	}

	var data struct {
		Result  bool   `json:"result"`
		Txid    string `json:"txid"`
		Message string `json:"message"`
		Code    string `json:"code"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", fmt.Errorf("decoding broadcast tx result: %w", err)
	}
	if !data.Result {
		return "", fmt.Errorf("broadcast result %s: %s", data.Code, data.Message)
	}

	return data.Txid, nil
}

func (r *Client) USDTBalance(ctx context.Context, addr string) (uint, error) {
	body := map[string]any{
		"owner_address":     addr,
		"contract_address":  usdtContractAddr(r.Net),
		"function_selector": "balanceOf(address)",
		"parameter":         abiEncodeAddr(addr),
		"call_value":        0,
		"visible":           true,
	}
	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost, r.url("/walletsolidity/triggerconstantcontract"),
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("TRON-PRO-API-KEY", r.apikey)

	resp, err := r.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("getting usdt balance of addr %s: %w", addr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("getting usdt balance of addr %s: %s", addr, resp.Status)
	}

	var data TriggerConstContract
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return 0, fmt.Errorf("decoding usdt balance of addr %s: %w", addr, err)
	}
	if !data.Result.Result {
		return 0, fmt.Errorf("usdt balance result of addr %s: %v", addr, data.Result.Result)
	}
	if len(data.ConstantResult) == 0 {
		return 0, fmt.Errorf("usdt balance result of addr %s: constantresult was empty", addr)
	}

	cRes := data.ConstantResult[0]
	balance, err := strconv.ParseUint(cRes, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing usdt balance of addr %s: %w", addr, err)
	}

	return uint(balance), nil
}

func (r *Client) SendUSDT(from, to string, amt uint) (*Tx, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := map[string]any{
		"owner_address":     from,
		"contract_address":  usdtContractAddr(r.Net),
		"function_selector": "transfer(address,uint256)",
		"parameter":         abiEncodeSend(to, amt),
		"visible":           true,
		"fee_limit":         10_000_000, // 10 usdt
	}
	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost, r.url("/wallet/triggersmartcontract"),
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("TRON-PRO-API-KEY", r.apikey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending usdt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sending usdt: %s", resp.Status)
	}

	var data TriggerSmartContract
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("decoding usdt tx: %w", err)
	}
	if !data.Result.Result {
		return nil, fmt.Errorf("sending usdt result: %v", data.Result.Result)
	}

	return &data.Transaction, nil
}

func abiEncodeSend(addr string, amt uint) string {
	encodedAddr := abiEncodeAddr(addr)
	encodedAmt := abiEncodeUint(amt)
	return encodedAddr + encodedAmt
}

func abiEncodeAddr(addr string) string {
	addrBytes := base58.Decode(addr)

	// extract address, leaves 20 bytes
	addrBytes = addrBytes[1:21]

	// pad to total of 32 bytes
	padding := [12]byte{}
	addrBytes = append(padding[:], addrBytes...)

	// done
	return hex.EncodeToString(addrBytes)
}

func abiEncodeUint(v uint) string {
	value := big.NewInt(int64(v))
	b := value.Bytes()
	out := make([]byte, 32)
	copy(out[32-len(b):], b)
	return hex.EncodeToString(out)
}

func (r *Client) url(path string) string {
	switch r.Net {
	case chain.Mainnet:
		return "https://api.trongrid.io" + path
	case chain.Testnet:
		return "https://api.shasta.trongrid.io" + path
	}
	return "unreachable" // compiler is too dumb
}
