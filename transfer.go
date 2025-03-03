package main

import "C"
import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dogecoinw/doged/btcec"
	"github.com/dogecoinw/doged/btcutil"
	"github.com/dogecoinw/doged/chaincfg"
	"github.com/dogecoinw/doged/chaincfg/chainhash"
	"github.com/dogecoinw/doged/rpcclient"
	"github.com/dogecoinw/doged/txscript"
	"github.com/dogecoinw/doged/wire"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
)

const (
	TX_FEE_AMOUNT = 2000000
	TX_FEE        = float64(TX_FEE_AMOUNT / 1e8)
	BASE_AMOUNT   = 100000
	FEE_AMOUNT    = 2000000

	DEPLOY_AMOUNT   = 500000000
	TIP_AMOUNT      = 10000000
	FEE_ADDRESS     = "DJ9wVHBFnbcZUtfWdHWPEnijdxz1CABPUY"
	TIP_FEE_ADDRESS = "DSPAZ6cZC7ShL63UFKPgs4vBGrbpHBwWQG"
)

type Transfer struct {
	UtxoClient  *rpcclient.Client
	PrivateUtxo *btcec.PrivateKey
	Address     btcutil.Address
}

func NewTransfer(utxoClient *rpcclient.Client, private *btcec.PrivateKey) *Transfer {

	address, _ := btcutil.NewAddressPubKeyHash(btcutil.Hash160(private.PubKey().SerializeCompressed()), &chaincfg.MainNetParams)

	return &Transfer{
		UtxoClient:  utxoClient,
		PrivateUtxo: private,
		Address:     address,
	}
}

func (t *Transfer) TransferInscription(pump []*Inscription) (map[string]*wire.MsgTx, error) {

	inscript := map[btcutil.Address][]byte{}
	inscriptAddress := make([]btcutil.Address, 0)
	for _, v := range pump {
		va, _ := json.Marshal(v)
		println(string(va))
		var script []byte
		if v.Op == "deploy" {
			script, _ = InscriptionDeployScript(v, t.PrivateUtxo.PubKey().SerializeCompressed())
		} else {
			script, _ = InscriptionScript(v, t.PrivateUtxo.PubKey().SerializeCompressed())
		}

		addr, _ := btcutil.NewAddressScriptHash(script, &chaincfg.MainNetParams)
		println(addr.String())
		inscript[addr] = script
		inscriptAddress = append(inscriptAddress, addr)
	}

	multiTx, err := t.CreateMultiOutScript(inscriptAddress)
	if err != nil {
		return nil, err
	}

	_, err = t.UtxoClient.SendRawTransaction(multiTx, true)
	if err != nil {
		return nil, err
	}

	inscriptTx := make(map[btcutil.Address]*wire.MsgTx)
	nameTx := make(map[string]*wire.MsgTx)
	for i, addr := range inscriptAddress {
		time.Sleep(1 * time.Second)
		tx := wire.NewMsgTx(wire.TxVersion)

		script := inscript[addr]

		// add input
		txHash, _ := chainhash.NewHashFromStr(multiTx.TxHash().String())
		outPoint := wire.NewOutPoint(txHash, uint32(i))
		txIn := wire.NewTxIn(outPoint, nil, nil)
		tx.AddTxIn(txIn)

		decodedAddr, err := btcutil.DecodeAddress(pump[i].HolderAddress, &chaincfg.MainNetParams)
		if err != nil {
			return nil, err
		}

		scriptAdd, err := txscript.PayToAddrScript(decodedAddr)
		if err != nil {
			return nil, err
		}

		// add output
		txOut := wire.NewTxOut(BASE_AMOUNT, scriptAdd)
		tx.AddTxOut(txOut)

		if pump[i].Op == "deploy" {

			depolyAddr, err := btcutil.DecodeAddress(FEE_ADDRESS, &chaincfg.MainNetParams)
			if err != nil {
				return nil, err
			}

			scriptDepolyAdd, err := txscript.PayToAddrScript(depolyAddr)
			if err != nil {
				return nil, err
			}

			// add output
			txOut := wire.NewTxOut(DEPLOY_AMOUNT, scriptDepolyAdd)
			tx.AddTxOut(txOut)
		}

		depolyTipAddr, err := btcutil.DecodeAddress(TIP_FEE_ADDRESS, &chaincfg.MainNetParams)
		if err != nil {
			return nil, err
		}

		scriptTipAdd, err := txscript.PayToAddrScript(depolyTipAddr)
		if err != nil {
			return nil, err
		}

		// add output
		txTipOut := wire.NewTxOut(TIP_AMOUNT, scriptTipAdd)
		tx.AddTxOut(txTipOut)

		// sign transaction
		signature, err := txscript.RawTxInSignature(tx, 0, script, txscript.SigHashAll, t.PrivateUtxo)
		if err != nil {
			return nil, err
		}

		signatureScript := txscript.NewScriptBuilder()
		signatureScript.AddOp(txscript.OP_10).AddOp(txscript.OP_FALSE).AddData(signature)
		signatureScript.AddData(script)
		sigScript, err := signatureScript.Script()
		if err != nil {
			return nil, err
		}

		tx.TxIn[0].SignatureScript = sigScript
		_, err = t.UtxoClient.SendRawTransaction(tx, true)
		if err != nil {
			return nil, err
		}

		inscriptTx[addr] = tx
		nameTx[pump[i].Name] = tx

		var signedTx bytes.Buffer
		err = tx.Serialize(&signedTx)
		if err != nil {
			return nil, err
		}

		hexSignedTx := hex.EncodeToString(signedTx.Bytes())
		fmt.Println(hexSignedTx)
	}

	inscriptTxStatus := make(map[btcutil.Address]*wire.MsgTx)
	for key, value := range inscriptTx {
		inscriptTxStatus[key] = value
	}

	return nameTx, nil
}

func (t *Transfer) InscriptionScript(pump *Inscription) ([]byte, error) {

	builder := txscript.NewScriptBuilder()
	data := make(map[string]interface{})
	data["p"] = pump.P
	data["op"] = pump.Op
	data["pair_id"] = pump.PairId
	data["tick0_id"] = pump.Tick0Id
	data["amt0"] = pump.Amt0
	data["amt1_min"] = pump.Amt1Min
	data["doge"] = pump.Doge
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	//create redeem script
	builder.AddOp(txscript.OP_1).AddData(t.PrivateUtxo.PubKey().SerializeCompressed()).AddOp(txscript.OP_1)
	builder.AddOp(txscript.OP_CHECKMULTISIGVERIFY)
	builder.AddData([]byte("ord")).AddData([]byte("text/plain;charset=utf-8")).AddData(jsonData)
	builder.AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP)

	// redeem script is the script program in the format of []byte
	redeemScript, err := builder.Script()
	if err != nil {
		return nil, err
	}

	return redeemScript, nil
}

func (t *Transfer) InscriptionDeployScript(pump *Inscription) ([]byte, error) {

	builder := txscript.NewScriptBuilder()
	data := make(map[string]interface{})
	data["p"] = pump.P
	data["op"] = pump.Op
	data["tick"] = pump.Tick
	data["symbol"] = pump.Symbol
	data["name"] = pump.Name
	data["amt"] = pump.Amt
	data["doge"] = pump.Doge
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	//create redeem script
	builder.AddOp(txscript.OP_1).AddData(t.PrivateUtxo.PubKey().SerializeCompressed()).AddOp(txscript.OP_1)
	builder.AddOp(txscript.OP_CHECKMULTISIGVERIFY)
	builder.AddData([]byte("ord")).AddData([]byte("text/plain;charset=utf-8")).AddData(jsonData)
	builder.AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP).AddOp(txscript.OP_DROP)

	// redeem script is the script program in the format of []byte
	redeemScript, err := builder.Script()
	if err != nil {
		return nil, err
	}

	return redeemScript, nil
}

func (t *Transfer) CreateMultiOutScript(address []btcutil.Address) (*wire.MsgTx, error) {

	fee := TX_FEE * float64(len(address))

	utxos, err := GetUtxo(t.Address.String(), fmt.Sprintf("%f", fee))
	if err != nil {
		return nil, err
	}

	// create a new transaction
	redeemTx := wire.NewMsgTx(wire.TxVersion)

	// add input
	utxoAmountSum := 0.0
	for _, utxo := range utxos.Utxo {
		txHash, _ := chainhash.NewHashFromStr(utxo.Txid)
		outPoint := wire.NewOutPoint(txHash, uint32(utxo.Vout))
		txIn := wire.NewTxIn(outPoint, nil, nil)
		redeemTx.AddTxIn(txIn)
		utxoAmountSum += utxo.Value
	}

	// add output
	for _, addr := range address {
		println("CreateMultiOutScript", addr.String())
		// create a new output
		script, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
		txOut := wire.NewTxOut(TX_FEE_AMOUNT, script)
		redeemTx.AddTxOut(txOut)
	}

	script, err := txscript.PayToAddrScript(t.Address)
	if err != nil {
		return nil, err
	}

	am := int64(utxoAmountSum*(1e8)) - int64(fee*(1e8)) - FEE_AMOUNT
	if am < 0 {
		return nil, fmt.Errorf("utxo amount not enough")
	}

	txOut := wire.NewTxOut(am, script)
	redeemTx.AddTxOut(txOut)

	// sign transaction
	for i, _ := range redeemTx.TxIn {
		signature, err := txscript.SignatureScript(redeemTx, i, script, txscript.SigHashAll, t.PrivateUtxo, true)
		if err != nil {
			return nil, err
		}
		redeemTx.TxIn[i].SignatureScript = signature
	}

	return redeemTx, nil
}

func GetUtxo(address, amt string) (*UtxoResponse, error) {
	url := "https://utxo.unielon.com/utxo"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("address", address)
	_ = writer.WriteField("amount", amt)
	_ = writer.WriteField("count", "10")
	_ = writer.WriteField("small_change", "1")
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	utxo := &UtxoResponse{
		Amount: 0,
		Utxo:   make([]*UtxoContent, 0),
	}

	err = json.Unmarshal(body, &utxo)
	if err != nil {
		return nil, err
	}

	return utxo, nil
}
