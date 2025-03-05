package main

import (
	"github.com/dogecoinw/doged/btcutil"
	"github.com/dogecoinw/doged/chaincfg"
	"github.com/dogecoinw/doged/rpcclient"
	"testing"
)

// deploy
func TestDeploy(t *testing.T) {

	LoadConfig(&cfg, "")

	wifStr := ""
	wif, err := btcutil.DecodeWIF(wifStr)
	if err != nil {
		return
	}

	holderAddress, _ := btcutil.NewAddressPubKeyHash(btcutil.Hash160(wif.PrivKey.PubKey().SerializeCompressed()), &chaincfg.MainNetParams)

	connCfg := &rpcclient.ConnConfig{
		Host:         cfg.Rpc,
		Endpoint:     "ws",
		User:         cfg.UserName,
		Pass:         cfg.PassWord,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}

	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	rpcClient, _ := rpcclient.New(connCfg, nil)

	transfer := NewTransfer(rpcClient, wif.PrivKey)

	inscriptions := make([]*Inscription, 0)

	inscriptions = append(inscriptions, &Inscription{
		P:             "pump",
		Op:            "deploy",
		Tick:          "WDOGE(WRAPPED-DOGE)",
		Symbol:        "symbol",
		Name:          "name",
		Logo:          "path",
		Amt:           "0",
		Doge:          0,
		HolderAddress: holderAddress.String(),
	})

	txs, err := transfer.TransferInscription(inscriptions)
	if err != nil {
		println(err.Error())
		return
	}

	for _, v := range txs {
		println(v.TxHash().String())
	}
}

func TestTrade(t *testing.T) {

	LoadConfig(&cfg, "")

	wifStr := ""
	wif, err := btcutil.DecodeWIF(wifStr)
	if err != nil {
		return
	}

	holderAddress, _ := btcutil.NewAddressPubKeyHash(btcutil.Hash160(wif.PrivKey.PubKey().SerializeCompressed()), &chaincfg.MainNetParams)

	// After deployment
	pairId := ""
	tickId := ""

	connCfg := &rpcclient.ConnConfig{
		Host:         cfg.Rpc,
		Endpoint:     "ws",
		User:         cfg.UserName,
		Pass:         cfg.PassWord,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}

	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	rpcClient, _ := rpcclient.New(connCfg, nil)

	transfer := NewTransfer(rpcClient, wif.PrivKey)

	inscriptions := make([]*Inscription, 0)

	inscriptions = append(inscriptions, &Inscription{
		P:             "pump",
		Op:            "trade",
		PairId:        pairId,
		Tick0Id:       tickId,
		Amt0:          "100000000",
		Amt1Min:       "0",
		Doge:          0,
		HolderAddress: holderAddress.String(),
	})

	txs, err := transfer.TransferInscription(inscriptions)
	if err != nil {
		println(err.Error())
		return
	}

	for _, v := range txs {
		println(v.TxHash().String())
	}

}
