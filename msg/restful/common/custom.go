package common

import (
	. "IPT/common"
	"IPT/core/ledger"
	tx "IPT/core/transaction"
	"IPT/core/transaction/payload"
	. "IPT/common/errors"
	Err "IPT/msg/restful/error"
	. "IPT/msg/rpc"
	"bytes"
	"encoding/json"
)

const AttributeMaxLen = 252

type RecordData struct {
	Hash string                 `json:Hash`
	Data map[string]interface{} `json:Data`
}

func SendRecordTransaction(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	var recordData []byte
	reqRecordData, ok := cmd["RecordData"].(map[string]interface{})
	if !ok {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	recordData, err := json.Marshal(reqRecordData)
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	if recordData == nil {
		return resp
	}
	recordType := "record"
	recordTx, _ := tx.NewRecordTransaction(recordType, recordData)

	hash := recordTx.Hash()
	resp["Result"] = BytesToHexString(hash.ToArrayReverse())
	if errCode := VerifyAndSendTx(recordTx); errCode != ErrNoError {
		resp["Error"] = int64(errCode)
		return resp
	}
	return resp
}

func GetRecordByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)

	str := cmd["Hash"].(string)
	bys, err := HexStringToBytesReverse(str)
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	var hash Uint256
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		resp["Error"] = Err.INVALID_TRANSACTION
		return resp
	}
	tx, err := ledger.DefaultLedger.Store.GetTransaction(hash)
	if err != nil {
		resp["Error"] = Err.UNKNOWN_RECORD
		return resp
	}
	recordinfo := tx.Payload.(*payload.Record)

	tmp := &RecordData{}
	if err := json.Unmarshal(recordinfo.RecordData, tmp); err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	resp["Result"] = tmp
	return resp
}

func GetRecordByFileHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)

	str := cmd["Hash"].(string)

	txhash, err := ledger.DefaultLedger.Store.GetRecord(str)
	if err != nil {
		resp["Error"] = Err.UNKNOWN_RECORD
		return resp
	}
	tx, err := ledger.DefaultLedger.Store.GetTransaction(txhash)
	if err != nil {
		resp["Error"] = Err.UNKNOWN_TRANSACTION
		return resp
	}
	recordinfo := tx.Payload.(*payload.Record)

	tmp := &RecordData{}
	if err := json.Unmarshal(recordinfo.RecordData, tmp); err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	resp["Result"] = tmp
	return resp
}
