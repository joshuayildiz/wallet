package trongrid

// todo: check what this should look like
type Block struct {
	BlockID     string `json:"blockID"`
	BlockHeader struct {
		RawData struct {
			Number         uint   `json:"number"`
			ParentHash     string `json:"parentHash"`
			TxTrieRoot     string `json:"txTrieRoot"`
			WitnessAddress string `json:"witness_address"`
		} `json:"raw_data"`
	} `json:"block_header"`
	Transactions []Tx `json:"transactions"`
}

type Tx struct {
	RawData struct {
		Contract      []Contract `json:"contract"`
		Expiration    uint       `json:"expiration"`
		RefBlockBytes string     `json:"ref_block_bytes"`
		RefBlockHash  string     `json:"ref_block_hash"`
		Timestamp     uint       `json:"timestamp"`
		FeeLimit      uint       `json:"fee_limit"`
	} `json:"raw_data"`
	RawDataHex string   `json:"raw_data_hex"`
	TxID       string   `json:"txID"`
	Visible    bool     `json:"visible"`
	Signature  []string `json:"signature"`
}

type Contract struct {
	Parameter struct {
		TypeURL string `json:"type_url"`
		Value   struct {
			Amount          int    `json:"amount"`
			OwnerAddress    string `json:"owner_address"`
			ToAddress       string `json:"to_address"`
			Data            string `json:"data"`
			ContractAddress string `json:"contract_address"`
		} `json:"value"`
	} `json:"parameter"`
	Type string `json:"type"`
}

type TxInfo struct {
	BlockNumber     int      `json:"blockNumber"`
	BlockTimeStamp  int64    `json:"blockTimeStamp"`
	ContractResult  []string `json:"contractResult"`
	ContractAddress string   `json:"contract_address"`
	Fee             int      `json:"fee"`
	ID              string   `json:"id"`
	Log             []struct {
		Address string   `json:"address"`
		Data    string   `json:"data"`
		Topics  []string `json:"topics"`
	} `json:"log"`
	Receipt struct {
		EnergyFee         int    `json:"energy_fee"`
		EnergyUsageTotal  int    `json:"energy_usage_total"`
		NetUsage          int    `json:"net_usage"`
		OriginEnergyUsage int    `json:"origin_energy_usage"`
		Result            string `json:"result"`
	} `json:"receipt"`
}

type TriggerConstContract struct {
	Result struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Result  bool   `json:"result"`
	} `json:"result"`
	EnergyUsed     int      `json:"energy_used"`
	ConstantResult []string `json:"constant_result"`
	Transaction    Tx       `json:"transaction"`
}

type TriggerSmartContract struct {
	Result struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Result  bool   `json:"result"`
	} `json:"result"`
	Transaction Tx `json:"transaction"`
}
