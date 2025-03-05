package main

type UtxoContent struct {
	Txid    string  `json:"txid"`
	Vout    int     `json:"vout"`
	Address string  `json:"address"`
	Value   float64 `json:"value"`
}

type UtxoResponse struct {
	Amount float64        `json:"amount"`
	Utxo   []*UtxoContent `json:"utxo"`
}

type Inscription struct {
	P             string `json:"p"`
	Op            string `json:"op"`
	PairId        string `json:"pair_id"`
	Tick0Id       string `json:"tick0_id"`
	Amt0          string `json:"amt0"`
	Amt1Min       string `json:"amt1_min"`
	Doge          int    `json:"doge"`
	Tick          string `json:"tick"`
	Symbol        string `json:"symbol"`
	Logo          string `json:"logo"`
	Name          string `json:"name"`
	Amt           string `json:"amt"`
	HolderAddress string `json:"holder_address"`
}
