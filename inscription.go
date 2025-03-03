package main

import (
	"encoding/json"
	"github.com/dogecoinw/doged/txscript"
)

func InscriptionDeployScript(pump *Inscription, pubkey []byte) ([]byte, error) {

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
	builder.AddOp(txscript.OP_1).AddData(pubkey).AddOp(txscript.OP_1)
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

func InscriptionScript(pump *Inscription, pubkey []byte) ([]byte, error) {

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
	builder.AddOp(txscript.OP_1).AddData(pubkey).AddOp(txscript.OP_1)
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
