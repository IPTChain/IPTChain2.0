package common

import (
	"IPT/account"
	"IPT/common"
	. "IPT/common/errors"
	"IPT/contracts/types"
	"IPT/core/code"
	"IPT/core/contract"
	"IPT/core/signature"
	"IPT/core/transaction"
	Err "IPT/msg/restful/error"
	rpc "IPT/msg/rpc"
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
)

func newContractContextWithoutProgramHashes(data signature.SignableData) *contract.ContractContext {
	return &contract.ContractContext{
		Data:       data,
		Codes:      make([][]byte, 1),
		Parameters: make([][][]byte, 1),
	}
}

func signTransaction(signer *account.Account, tx *transaction.Transaction) error {
	signature, err := signature.SignBySigner(tx, signer)
	if err != nil {
		fmt.Println("SignBySigner failed")
		return err
	}
	transactionContract, err := contract.CreateSignatureContract(signer.PubKey())
	if err != nil {
		fmt.Println("CreateSignatureContract failed")
		return err
	}
	transactionContractContext := newContractContextWithoutProgramHashes(tx)
	if err := transactionContractContext.AddContract(transactionContract, signer.PubKey(), signature); err != nil {
		fmt.Println("AddContract failed")
		return err
	}
	tx.SetPrograms(transactionContractContext.GetPrograms())
	return nil
}
func makeDeployContractTransaction(signer *account.Account, codeStr string) (string, error) {
	c, _ := common.HexStringToBytes(codeStr)
	fc := &code.FunctionCode{
		Code:           c,
		ParameterTypes: []contract.ContractParameterType{contract.ByteArray, contract.ByteArray},
		ReturnType:     contract.ContractParameterType(contract.Object),
	}
	fc.CodeHash()

	tx, err := transaction.NewDeployTransaction(fc, signer.ProgramHash, "IPT", "1.0", "IPT user", "user@onchain.com", "test uint", types.LangType(byte(0)))
	if err != nil {
		return "Deploy contracts fail!", err
	}
	txAttr := transaction.NewTxAttribute(transaction.Nonce, []byte(strconv.FormatInt(rand.Int63(), 10)))
	tx.Attributes = make([]*transaction.TxAttribute, 0)
	tx.Attributes = append(tx.Attributes, &txAttr)

	var buffer bytes.Buffer
	if err := tx.Serialize(&buffer); err != nil {
		fmt.Println("serialize registtransaction failed")
		return "", err
	}
	return hex.EncodeToString(buffer.Bytes()), nil
}

func makeInvokeTransaction(signer *account.Account, paramsStr, codeHashStr string) (string, error) {
	p, _ := common.HexStringToBytes(paramsStr)
	hash, _ := common.HexStringToBytesReverse(codeHashStr)
	p = append(p, 0x69)
	p = append(p, hash...)
	codeHash := common.BytesToUint160(hash)
	transactionContract, err := contract.CreateSignatureContract(signer.PubKey())
	if err != nil {
		fmt.Println("CreateSignatureContract failed")
		return "", err
	}

	tx, err := transaction.NewInvokeTransaction(p, codeHash, transactionContract.ProgramHash)
	if err != nil {
		return "Invoke contracts fail!", err
	}
	txAttr := transaction.NewTxAttribute(transaction.Nonce, []byte(strconv.FormatInt(rand.Int63(), 10)))
	tx.Attributes = make([]*transaction.TxAttribute, 0)
	tx.Attributes = append(tx.Attributes, &txAttr)

	if err := signTransaction(signer, tx); err != nil {
		fmt.Println("sign transfer transaction failed")
		return "", err
	}
	var buffer bytes.Buffer
	if err := tx.Serialize(&buffer); err != nil {
		fmt.Println("serialize registtransaction failed")
		return "", err
	}
	return hex.EncodeToString(buffer.Bytes()), nil
}

func DeployContract(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	fmt.Printf("%+v", cmd)
	str, ok := cmd["Data"].(string)
	if !ok {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}

	programHashString, ok := cmd["ProgramHash"].(string)
	if !ok {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}

	programHashBytes, err := common.HexStringToBytesReverse(programHashString)
	if err != nil {
		fmt.Println("HexStringToBytes not found")
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}

	programHash, err := common.Uint160ParseFromBytes(programHashBytes)
	if err != nil {
		fmt.Println("Uint160ParseFromBytes not found")
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}

	signer := rpc.Wallet.GetAccountByProgramHash(programHash)

	if signer == nil {
		fmt.Println("signer not found")
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}

	txHex, err := makeDeployContractTransaction(signer, str)
	if err != nil {
		fmt.Println(err.Error())
	}

	bys, err := common.HexStringToBytes(txHex)
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	var txn transaction.Transaction
	if err := txn.Deserialize(bytes.NewReader(bys)); err != nil {
		resp["Error"] = Err.INVALID_TRANSACTION
		return resp
	}
	if txn.TxType != transaction.DeployCode {
		resp["Error"] = Err.INVALID_TRANSACTION
		return resp
	}
	var hash common.Uint256
	hash = txn.Hash()
	if errCode := rpc.VerifyAndSendTx(&txn); errCode != ErrNoError {
		resp["Error"] = int64(errCode)
		return resp
	}
	resp["Result"] = common.BytesToHexString(hash.ToArrayReverse())

	return resp
}

func InvokeContract(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	str, ok := cmd["Data"].(string)
	if !ok {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}

	param1, err := getParam(cmd["P1"])
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	param2, err := getParam(cmd["P2"])
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	param3, err := getParam(cmd["P3"])
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}

	param := generateParam(param3)
	param += generateParam(param2)
	param += generateParam(param1)

	programHashString, ok := cmd["ProgramHash"].(string)
	if !ok {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}

	programHashBytes, err := common.HexStringToBytesReverse(programHashString)
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}

	programHash, err := common.Uint160ParseFromBytes(programHashBytes)
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}

	signer := rpc.Wallet.GetAccountByProgramHash(programHash)

	if signer == nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}

	txHex, err := makeInvokeTransaction(signer, param, str)
	if err != nil {
		fmt.Println(err.Error())
	}

	bys, err := common.HexStringToBytes(txHex)
	if err != nil {
		resp["Error"] = Err.INVALID_PARAMS
		return resp
	}
	var txn transaction.Transaction
	if err := txn.Deserialize(bytes.NewReader(bys)); err != nil {
		fmt.Printf(err.Error())
		resp["Error"] = Err.INVALID_TRANSACTION
		return resp
	}
	if txn.TxType != transaction.InvokeCode {
		fmt.Printf("%+v", txn)
		resp["Error"] = Err.INVALID_TRANSACTION
		return resp
	}
	var hash common.Uint256
	hash = txn.Hash()
	if errCode := rpc.VerifyAndSendTx(&txn); errCode != ErrNoError {
		resp["Error"] = int64(errCode)
		return resp
	}
	resp["Result"] = common.BytesToHexString(hash.ToArrayReverse())

	return resp
}

func getParam(v interface{}) (string, error) {
	switch v.(type) {
	case string:
		s := v.(string)
		return common.BytesToHexString([]byte(s)), nil
	case int:
		s := v.(int)
		return common.BytesToHexString(common.IntToBytes(s)), nil
	}
	return "", nil
}

func generateParam(P string) string {
	var parString string
	lenP := len(P)
	if lenP == 0 {
		parString += "0100"
	} else {
		parString += common.BytesToHexString(common.ByteToBytes(lenP/2)) + P
	}
	return parString
}
