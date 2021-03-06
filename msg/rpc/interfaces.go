package rpc

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"IPT/account"
	. "IPT/common"
	"IPT/common/config"
	. "IPT/common/errors"
	"IPT/common/log"
	"IPT/core/ledger"
	"IPT/core/signature"
	tx "IPT/core/transaction"
	"IPT/sdk"

	"github.com/mitchellh/go-homedir"
)

const (
	RANDBYTELEN = 4
)

func TransArryByteToHexString(ptx *tx.Transaction) *Transactions {

	trans := new(Transactions)
	trans.TxType = ptx.TxType
	trans.PayloadVersion = ptx.PayloadVersion
	trans.Payload = TransPayloadToHex(ptx.Payload)

	n := 0
	trans.Attributes = make([]TxAttributeInfo, len(ptx.Attributes))
	for _, v := range ptx.Attributes {
		trans.Attributes[n].Usage = v.Usage
		trans.Attributes[n].Data = BytesToHexString(v.Data)
		n++
	}

	n = 0
	trans.UTXOInputs = make([]UTXOTxInputInfo, len(ptx.UTXOInputs))
	for _, v := range ptx.UTXOInputs {
		trans.UTXOInputs[n].ReferTxID = BytesToHexString(v.ReferTxID.ToArrayReverse())
		trans.UTXOInputs[n].ReferTxOutputIndex = v.ReferTxOutputIndex
		n++
	}

	n = 0
	trans.BalanceInputs = make([]BalanceTxInputInfo, len(ptx.BalanceInputs))
	for _, v := range ptx.BalanceInputs {
		trans.BalanceInputs[n].AssetID = BytesToHexString(v.AssetID.ToArrayReverse())
		trans.BalanceInputs[n].Value = v.Value.String()
		trans.BalanceInputs[n].ProgramHash = BytesToHexString(v.ProgramHash.ToArrayReverse())
		n++
	}

	n = 0
	trans.Outputs = make([]TxoutputInfo, len(ptx.Outputs))
	for _, v := range ptx.Outputs {
		trans.Outputs[n].AssetID = BytesToHexString(v.AssetID.ToArrayReverse())
		trans.Outputs[n].Value = v.Value.String()
		address, _ := v.ProgramHash.ToAddress()
		trans.Outputs[n].Address = address
		n++
	}

	n = 0
	trans.Programs = make([]ProgramInfo, len(ptx.Programs))
	for _, v := range ptx.Programs {
		trans.Programs[n].Code = BytesToHexString(v.Code)
		trans.Programs[n].Parameter = BytesToHexString(v.Parameter)
		n++
	}

	n = 0
	trans.AssetOutputs = make([]TxoutputMap, len(ptx.AssetOutputs))
	for k, v := range ptx.AssetOutputs {
		trans.AssetOutputs[n].Key = k
		trans.AssetOutputs[n].Txout = make([]TxoutputInfo, len(v))
		for m := 0; m < len(v); m++ {
			trans.AssetOutputs[n].Txout[m].AssetID = BytesToHexString(v[m].AssetID.ToArrayReverse())
			trans.AssetOutputs[n].Txout[m].Value = v[m].Value.String()
			trans.AssetOutputs[n].Txout[m].Address = BytesToHexString(v[m].ProgramHash.ToArray())
		}
		n += 1
	}

	n = 0
	trans.AssetInputAmount = make([]AmountMap, len(ptx.AssetInputAmount))
	for k, v := range ptx.AssetInputAmount {
		trans.AssetInputAmount[n].Key = k
		trans.AssetInputAmount[n].Value = v
		n += 1
	}

	n = 0
	trans.AssetOutputAmount = make([]AmountMap, len(ptx.AssetOutputAmount))
	for k, v := range ptx.AssetOutputAmount {
		trans.AssetInputAmount[n].Key = k
		trans.AssetInputAmount[n].Value = v
		n += 1
	}

	mHash := ptx.Hash()
	trans.Hash = BytesToHexString(mHash.ToArrayReverse())

	return trans
}
func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}
func getBestBlockHash(params []interface{}) map[string]interface{} {
	hash := ledger.DefaultLedger.Blockchain.CurrentBlockHash()
	return IPTRpc(BytesToHexString(hash.ToArrayReverse()))
}

// Input JSON string examples for getblock method as following:
//   {"jsonrpc": "2.0", "method": "getblock", "params": [1], "id": 0}
//   {"jsonrpc": "2.0", "method": "getblock", "params": ["aabbcc.."], "id": 0}
func getBlock(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return IPTRpcNil
	}
	var err error
	var hash Uint256
	switch (params[0]).(type) {
	// block height
	case float64:
		index := uint32(params[0].(float64))
		hash, err = ledger.DefaultLedger.Store.GetBlockHash(index)
		if err != nil {
			return IPTRpcUnknownBlock
		}
	// block hash
	case string:
		str := params[0].(string)
		hex, err := HexStringToBytesReverse(str)
		if err != nil {
			return IPTRpcInvalidParameter
		}
		if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
			return IPTRpcInvalidTransaction
		}
	default:
		return IPTRpcInvalidParameter
	}

	block, err := ledger.DefaultLedger.Store.GetBlock(hash)
	if err != nil {
		return IPTRpcUnknownBlock
	}

	blockHead := &BlockHead{
		Version:          block.Blockdata.Version,
		PrevBlockHash:    BytesToHexString(block.Blockdata.PrevBlockHash.ToArrayReverse()),
		TransactionsRoot: BytesToHexString(block.Blockdata.TransactionsRoot.ToArrayReverse()),
		Timestamp:        block.Blockdata.Timestamp,
		Height:           block.Blockdata.Height,
		ConsensusData:    block.Blockdata.ConsensusData,
		NextBookKeeper:   BytesToHexString(block.Blockdata.NextBookKeeper.ToArrayReverse()),
		Program: ProgramInfo{
			Code:      BytesToHexString(block.Blockdata.Program.Code),
			Parameter: BytesToHexString(block.Blockdata.Program.Parameter),
		},
		Hash: BytesToHexString(hash.ToArrayReverse()),
	}

	trans := make([]*Transactions, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		trans[i] = TransArryByteToHexString(block.Transactions[i])
	}

	b := BlockInfo{
		Hash:         BytesToHexString(hash.ToArrayReverse()),
		BlockData:    blockHead,
		Transactions: trans,
	}
	return IPTRpc(b)
}

func getBlockCount(params []interface{}) map[string]interface{} {
	return IPTRpc(ledger.DefaultLedger.Blockchain.BlockHeight + 1)
}

// A JSON example for getblockhash method as following:
//   {"jsonrpc": "2.0", "method": "getblockhash", "params": [1], "id": 0}
func getBlockHash(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return IPTRpcNil
	}
	switch params[0].(type) {
	case float64:
		height := uint32(params[0].(float64))
		hash, err := ledger.DefaultLedger.Store.GetBlockHash(height)
		if err != nil {
			return IPTRpcUnknownBlock
		}
		return IPTRpc(BytesToHexString(hash.ToArrayReverse()))
	default:
		return IPTRpcInvalidParameter
	}
}

func getConnectionCount(params []interface{}) map[string]interface{} {
	return IPTRpc(node.GetConnectionCnt())
}

func getRawMemPool(params []interface{}) map[string]interface{} {
	txs := []*Transactions{}
	txpool := node.GetTxnPool(false)
	for _, t := range txpool {
		txs = append(txs, TransArryByteToHexString(t))
	}
	if len(txs) == 0 {
		return IPTRpcNil
	}
	return IPTRpc(txs)
}

// A JSON example for getrawtransaction method as following:
//   {"jsonrpc": "2.0", "method": "getrawtransaction", "params": ["transactioin hash in hex"], "id": 0}
func getRawTransaction(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return IPTRpcNil
	}
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := HexStringToBytesReverse(str)
		if err != nil {
			return IPTRpcInvalidParameter
		}
		var hash Uint256
		err = hash.Deserialize(bytes.NewReader(hex))
		if err != nil {
			return IPTRpcInvalidTransaction
		}
		tx, err := ledger.DefaultLedger.Store.GetTransaction(hash)
		if err != nil {
			return IPTRpcUnknownTransaction
		}
		tran := TransArryByteToHexString(tx)
		return IPTRpc(tran)
	default:
		return IPTRpcInvalidParameter
	}
}

// A JSON example for sendrawtransaction method as following:
//   {"jsonrpc": "2.0", "method": "sendrawtransaction", "params": ["raw transactioin in hex"], "id": 0}
func sendRawTransaction(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return IPTRpcNil
	}
	var hash Uint256
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := HexStringToBytes(str)
		if err != nil {
			return IPTRpcInvalidParameter
		}
		var txn tx.Transaction
		if err := txn.Deserialize(bytes.NewReader(hex)); err != nil {
			return IPTRpcInvalidTransaction
		}
		if txn.TxType != tx.InvokeCode && txn.TxType != tx.DeployCode &&
			txn.TxType != tx.TransferAsset && txn.TxType != tx.LockAsset &&
			txn.TxType != tx.RegisterAsset && txn.TxType != tx.IssueAsset &&
			txn.TxType != tx.BookKeeper {
			return IPTRpc("invalid transaction type")
		}
		hash = txn.Hash()
		if errCode := VerifyAndSendTx(&txn); errCode != ErrNoError {
			return IPTRpc(errCode.Error())
		}
	default:
		return IPTRpcInvalidParameter
	}
	return IPTRpc(BytesToHexString(hash.ToArrayReverse()))
}

func getTxout(params []interface{}) map[string]interface{} {
	//TODO
	return IPTRpcUnsupported
}

// A JSON example for submitblock method as following:
//   {"jsonrpc": "2.0", "method": "submitblock", "params": ["raw block in hex"], "id": 0}
func submitBlock(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return IPTRpcNil
	}
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, _ := HexStringToBytes(str)
		var block ledger.Block
		if err := block.Deserialize(bytes.NewReader(hex)); err != nil {
			return IPTRpcInvalidBlock
		}
		if err := ledger.DefaultLedger.Blockchain.AddBlock(&block); err != nil {
			return IPTRpcInvalidBlock
		}
		if err := node.Xmit(&block); err != nil {
			return IPTRpcInternalError
		}
	default:
		return IPTRpcInvalidParameter
	}
	return IPTRpcSuccess
}

func getNeighbor(params []interface{}) map[string]interface{} {
	addr, _ := node.GetNeighborAddrs()
	return IPTRpc(addr)
}

func getNodeState(params []interface{}) map[string]interface{} {
	n := NodeInfo{
		State:    uint(node.GetState()),
		Time:     node.GetTime(),
		Port:     node.GetPort(),
		ID:       node.GetID(),
		Version:  node.Version(),
		Services: node.Services(),
		Relay:    node.GetRelay(),
		Height:   node.GetHeight(),
		TxnCnt:   node.GetTxnCnt(),
		RxTxnCnt: node.GetRxTxnCnt(),
	}
	return IPTRpc(n)
}

func startConsensus(params []interface{}) map[string]interface{} {
	if err := dBFT.Start(); err != nil {
		return IPTRpcFailed
	}
	return IPTRpcSuccess
}

func stopConsensus(params []interface{}) map[string]interface{} {
	if err := dBFT.Halt(); err != nil {
		return IPTRpcFailed
	}
	return IPTRpcSuccess
}

func sendSampleTransaction(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return IPTRpcNil
	}
	var txType string
	switch params[0].(type) {
	case string:
		txType = params[0].(string)
	default:
		return IPTRpcInvalidParameter
	}

	issuer, err := account.NewAccount()
	if err != nil {
		return IPTRpc("Failed to create account")
	}
	admin := issuer

	rbuf := make([]byte, RANDBYTELEN)
	rand.Read(rbuf)
	switch string(txType) {
	case "perf":
		num := 1
		if len(params) == 2 {
			switch params[1].(type) {
			case float64:
				num = int(params[1].(float64))
			}
		}
		for i := 0; i < num; i++ {
			regTx := NewRegTx(BytesToHexString(rbuf), i, admin, issuer)
			SignTx(admin, regTx)
			VerifyAndSendTx(regTx)
		}
		return IPTRpc(fmt.Sprintf("%d transaction(s) was sent", num))
	default:
		return IPTRpc("Invalid transacion type")
	}
}

func setDebugInfo(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return IPTRpcInvalidParameter
	}
	switch params[0].(type) {
	case float64:
		level := params[0].(float64)
		if err := log.Log.SetDebugLevel(int(level)); err != nil {
			return IPTRpcInvalidParameter
		}
	default:
		return IPTRpcInvalidParameter
	}
	return IPTRpcSuccess
}

func getVersion(params []interface{}) map[string]interface{} {
	return IPTRpc(config.Version)
}

func uploadDataFile(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return IPTRpcNil
	}

	rbuf := make([]byte, 4)
	rand.Read(rbuf)
	tmpname := hex.EncodeToString(rbuf)

	str := params[0].(string)

	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return IPTRpcInvalidParameter
	}
	f, err := os.OpenFile(tmpname, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return IPTRpcIOError
	}
	defer f.Close()
	f.Write(data)

	refpath, err := AddFileIPFS(tmpname, true)
	if err != nil {
		return IPTRpcAPIError
	}

	return IPTRpc(refpath)

}

func regDataFile(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return IPTRpcNil
	}
	var hash Uint256
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := HexStringToBytes(str)
		if err != nil {
			return IPTRpcInvalidParameter
		}
		var txn tx.Transaction
		if err := txn.Deserialize(bytes.NewReader(hex)); err != nil {
			return IPTRpcInvalidTransaction
		}

		hash = txn.Hash()
		if errCode := VerifyAndSendTx(&txn); errCode != ErrNoError {
			return IPTRpcInternalError
		}
	default:
		return IPTRpcInvalidParameter
	}
	return IPTRpc(BytesToHexString(hash.ToArrayReverse()))
}

func catDataRecord(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return IPTRpcNil
	}
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		b, err := HexStringToBytesReverse(str)
		if err != nil {
			return IPTRpcInvalidParameter
		}
		var hash Uint256
		err = hash.Deserialize(bytes.NewReader(b))
		if err != nil {
			return IPTRpcInvalidTransaction
		}
		tx, err := ledger.DefaultLedger.Store.GetTransaction(hash)
		if err != nil {
			return IPTRpcUnknownTransaction
		}
		tran := TransArryByteToHexString(tx)
		info := tran.Payload.(*DataFileInfo)
		//ref := string(record.RecordData[:])
		return IPTRpc(info)
	default:
		return IPTRpcInvalidParameter
	}
}

func getDataFile(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return IPTRpcNil
	}
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := HexStringToBytesReverse(str)
		if err != nil {
			return IPTRpcInvalidParameter
		}
		var hash Uint256
		err = hash.Deserialize(bytes.NewReader(hex))
		if err != nil {
			return IPTRpcInvalidTransaction
		}
		tx, err := ledger.DefaultLedger.Store.GetTransaction(hash)
		if err != nil {
			return IPTRpcUnknownTransaction
		}

		tran := TransArryByteToHexString(tx)
		info := tran.Payload.(*DataFileInfo)

		err = GetFileIPFS(info.IPFSPath, info.Filename)
		if err != nil {
			return IPTRpcAPIError
		}
		//TODO: shoud return download address
		return IPTRpcSuccess
	default:
		return IPTRpcInvalidParameter
	}
}

var Wallet account.Client

func getWalletDir() string {
	home, _ := homedir.Dir()
	return home + "/.wallet/"
}

func createWallet(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return IPTRpcNil
	}
	var password []byte
	switch params[0].(type) {
	case string:
		password = []byte(params[0].(string))
	default:
		return IPTRpcInvalidParameter
	}
	walletDir := getWalletDir()
	if !FileExisted(walletDir) {
		err := os.MkdirAll(walletDir, 0755)
		if err != nil {
			return IPTRpcInternalError
		}
	}
	walletPath := walletDir + "wallet.dat"
	if FileExisted(walletPath) {
		return IPTRpcWalletAlreadyExists
	}
	_, err := account.Create(walletPath, password)
	if err != nil {
		return IPTRpcFailed
	}
	return IPTRpcSuccess
}

func openWallet(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return IPTRpcNil
	}
	var password []byte
	switch params[0].(type) {
	case string:
		password = []byte(params[0].(string))
	default:
		return IPTRpcInvalidParameter
	}
	resp := make(map[string]string)
	walletPath := getWalletDir() + "wallet.dat"
	if !FileExisted(walletPath) {
		resp["success"] = "false"
		resp["message"] = "wallet doesn't exist"
		return IPTRpc(resp)
	}
	wallet, err := account.Open(walletPath, password)
	if err != nil {
		resp["success"] = "false"
		resp["message"] = "password wrong"
		return IPTRpc(resp)
	}
	Wallet = wallet
	programHash, err := wallet.LoadStoredData("ProgramHash")
	if err != nil {
		resp["success"] = "false"
		resp["message"] = "wallet file broken"
		return IPTRpc(resp)
	}
	resp["success"] = "true"
	resp["message"] = BytesToHexString(programHash)
	return IPTRpc(resp)
}

func closeWallet(params []interface{}) map[string]interface{} {
	Wallet = nil
	return IPTRpcSuccess
}

func recoverWallet(params []interface{}) map[string]interface{} {
	if len(params) < 2 {
		return IPTRpcNil
	}
	var privateKey string
	var walletPassword string
	switch params[0].(type) {
	case string:
		privateKey = params[0].(string)
	default:
		return IPTRpcInvalidParameter
	}
	switch params[1].(type) {
	case string:
		walletPassword = params[1].(string)
	default:
		return IPTRpcInvalidParameter
	}
	walletDir := getWalletDir()
	if !FileExisted(walletDir) {
		err := os.MkdirAll(walletDir, 0755)
		if err != nil {
			return IPTRpcInternalError
		}
	}
	walletName := fmt.Sprintf("wallet-%s-recovered.dat", time.Now().Format("2006-01-02-15-04-05"))
	walletPath := walletDir + walletName
	if FileExisted(walletPath) {
		return IPTRpcWalletAlreadyExists
	}
	_, err := account.Recover(walletPath, []byte(walletPassword), privateKey)
	if err != nil {
		return IPTRpc("wallet recovery failed")
	}

	return IPTRpcSuccess
}

func getWalletKey(params []interface{}) map[string]interface{} {
	if Wallet == nil {
		return IPTRpc("open wallet first")
	}
	account, _ := Wallet.GetDefaultAccount()
	encodedPublickKey, _ := account.PublicKey.EncodePoint(true)
	resp := make(map[string]string)
	resp["PublicKey"] = BytesToHexString(encodedPublickKey)
	resp["PrivateKey"] = BytesToHexString(account.PrivateKey)
	resp["ProgramHash"] = BytesToHexString(account.ProgramHash.ToArrayReverse())

	return IPTRpc(resp)
}

func addAccount(params []interface{}) map[string]interface{} {
	if Wallet == nil {
		return IPTRpc("open wallet first")
	}
	account, err := Wallet.CreateAccount()
	if err != nil {
		return IPTRpc("create account error:" + err.Error())
	}

	if err := Wallet.CreateContract(account); err != nil {
		return IPTRpc("create contract error:" + err.Error())
	}

	address, err := account.ProgramHash.ToAddress()
	if err != nil {
		return IPTRpc("generate address error:" + err.Error())
	}

	return IPTRpc(address)
}

func deleteAccount(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return IPTRpcNil
	}
	var address string
	switch params[0].(type) {
	case string:
		address = params[0].(string)
	default:
		return IPTRpcInvalidParameter
	}
	if Wallet == nil {
		return IPTRpc("open wallet first")
	}
	programHash, err := ToScriptHash(address)
	if err != nil {
		return IPTRpc("invalid address:" + err.Error())
	}
	if err := Wallet.DeleteAccount(programHash); err != nil {
		return IPTRpc("Delete account error:" + err.Error())
	}
	if err := Wallet.DeleteContract(programHash); err != nil {
		return IPTRpc("Delete contract error:" + err.Error())
	}
	if err := Wallet.DeleteCoinsData(programHash); err != nil {
		return IPTRpc("Delete coins error:" + err.Error())
	}

	return IPTRpc(true)
}

func makeRegTxn(params []interface{}) map[string]interface{} {
	if len(params) < 2 {
		return IPTRpcNil
	}
	var assetName, assetValue string
	switch params[0].(type) {
	case string:
		assetName = params[0].(string)
	default:
		return IPTRpcInvalidParameter
	}
	switch params[1].(type) {
	case string:
		assetValue = params[1].(string)
	default:
		return IPTRpcInvalidParameter
	}
	if Wallet == nil {
		return IPTRpc("open wallet first")
	}

	regTxn, err := sdk.MakeRegTransaction(Wallet, assetName, assetValue)
	if err != nil {
		return IPTRpcInternalError
	}

	if errCode := VerifyAndSendTx(regTxn); errCode != ErrNoError {
		return IPTRpcInvalidTransaction
	}
	return IPTRpc(true)
}

func makeIssueTxn(params []interface{}) map[string]interface{} {
	if len(params) < 3 {
		return IPTRpcNil
	}
	var asset, value, address string
	switch params[0].(type) {
	case string:
		asset = params[0].(string)
	default:
		return IPTRpcInvalidParameter
	}
	switch params[1].(type) {
	case string:
		value = params[1].(string)
	default:
		return IPTRpcInvalidParameter
	}
	switch params[2].(type) {
	case string:
		address = params[2].(string)
	default:
		return IPTRpcInvalidParameter
	}
	if Wallet == nil {
		return IPTRpc("open wallet first")
	}
	tmp, err := HexStringToBytesReverse(asset)
	if err != nil {
		return IPTRpc("invalid asset ID")
	}
	var assetID Uint256
	if err := assetID.Deserialize(bytes.NewReader(tmp)); err != nil {
		return IPTRpc("invalid asset hash")
	}
	issueTxn, err := sdk.MakeIssueTransaction(Wallet, assetID, address, value)
	if err != nil {
		return IPTRpcInternalError
	}

	if errCode := VerifyAndSendTx(issueTxn); errCode != ErrNoError {
		return IPTRpcInvalidTransaction
	}

	return IPTRpc(true)
}

func sendToAddress(params []interface{}) map[string]interface{} {
	if len(params) < 3 {
		return IPTRpcNil
	}
	var asset, address, value, note string
	switch params[0].(type) {
	case string:
		asset = params[0].(string)
	default:
		return IPTRpcInvalidParameter
	}
	switch params[1].(type) {
	case string:
		address = params[1].(string)
	default:
		return IPTRpcInvalidParameter
	}
	switch params[2].(type) {
	case string:
		value = params[2].(string)
	default:
		return IPTRpcInvalidParameter
	}
	/*fmt.Println("-----------")
	fmt.Println(len(params))
	fmt.Println("-----------")*/
	if len(params) == 4 {
		switch params[3].(type) {
		case string:
			note = params[3].(string)
		default:
			return IPTRpcInvalidParameter
		}
	}
	/*fmt.Println("---------------")
	fmt.Println(node)
	fmt.Println("---------------")*/
	if Wallet == nil {
		return IPTRpc("error : wallet is not opened")
	}

	batchOut := sdk.BatchOut{
		Address: address,
		Value:   value,
		Note:    note,
	}
	tmp, err := HexStringToBytesReverse(asset)
	if err != nil {
		return IPTRpc("error: invalid asset ID")
	}
	var assetID Uint256
	if err := assetID.Deserialize(bytes.NewReader(tmp)); err != nil {
		return IPTRpc("error: invalid asset hash")
	}
	txn, err := sdk.MakeTransferTransaction(Wallet, assetID, batchOut)
	if err != nil {
		return IPTRpc("error: " + err.Error())
	}

	if errCode := VerifyAndSendTx(txn); errCode != ErrNoError {
		return IPTRpc("error: " + errCode.Error())
	}
	txHash := txn.Hash()
	return IPTRpc(BytesToHexString(txHash.ToArrayReverse()))
}

func createAccountForCust(params []interface{}) map[string]interface{} {
	if len(params) < 2 {
		return IPTRpcNil
	}
	var first, second string

	switch params[0].(type) {
	case string:
		first = params[0].(string)
	default:
		return IPTRpcInvalidParameter
	}
	switch params[1].(type) {
	case string:
		second = params[1].(string)
	default:
		return IPTRpcInvalidParameter
	}

	if first != second {
		return IPTRpc("error: Unmatched Password")
	}

	accountAddr, err := account.CreateAccountNotSave()
	if err != nil {
		return IPTRpc("error: " + err.Error())
	}

	address, _ := accountAddr.ProgramHash.ToAddress()
	publicKey, _ := accountAddr.PublicKey.EncodePoint(true)
	privateKey := accountAddr.PrivateKey
	a := AccountWalletInfo{
		Address:    address,
		PublicKey:  BytesToHexString(publicKey),
		PrivateKey: BytesToHexString(privateKey),
	}
	return IPTRpc(a)
}

func lockAsset(params []interface{}) map[string]interface{} {
	if len(params) < 3 {
		return IPTRpcNil
	}
	var asset, value string
	var height float64
	switch params[0].(type) {
	case string:
		asset = params[0].(string)
	default:
		return IPTRpcInvalidParameter
	}
	switch params[1].(type) {
	case string:
		value = params[1].(string)
	default:
		return IPTRpcInvalidParameter
	}
	switch params[2].(type) {
	case float64:
		height = params[2].(float64)
	default:
		return IPTRpcInvalidParameter
	}
	if Wallet == nil {
		return IPTRpc("error: invalid wallet instance")
	}

	accts := Wallet.GetAccounts()
	if len(accts) > 1 {
		return IPTRpc("error: does't support multi-addresses wallet locking asset")
	}

	tmp, err := HexStringToBytesReverse(asset)
	if err != nil {
		return IPTRpc("error: invalid asset ID")
	}
	var assetID Uint256
	if err := assetID.Deserialize(bytes.NewReader(tmp)); err != nil {
		return IPTRpc("error: invalid asset hash")
	}

	txn, err := sdk.MakeLockAssetTransaction(Wallet, assetID, value, uint32(height))
	if err != nil {
		return IPTRpc("error: " + err.Error())
	}

	txnHash := txn.Hash()
	if errCode := VerifyAndSendTx(txn); errCode != ErrNoError {
		return IPTRpc(errCode.Error())
	}
	return IPTRpc(BytesToHexString(txnHash.ToArrayReverse()))
}

func signMultisigTransaction(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return IPTRpcNil
	}
	var signedrawtxn string
	switch params[0].(type) {
	case string:
		signedrawtxn = params[0].(string)
	default:
		return IPTRpcInvalidParameter
	}

	rawtxn, _ := HexStringToBytes(signedrawtxn)
	var txn tx.Transaction
	txn.Deserialize(bytes.NewReader(rawtxn))
	if len(txn.Programs) <= 0 {
		return IPTRpc("missing the first signature")
	}

	found := false
	programHashes := txn.ParseTransactionCode()
	for _, hash := range programHashes {
		acct := Wallet.GetAccountByProgramHash(hash)
		if acct != nil {
			found = true
			sig, _ := signature.SignBySigner(&txn, acct)
			txn.AppendNewSignature(sig)
		}
	}
	if !found {
		return IPTRpc("error: no available account detected")
	}

	_, needsig, err := txn.ParseTransactionSig()
	if err != nil {
		return IPTRpc("error: " + err.Error())
	}
	if needsig == 0 {
		txnHash := txn.Hash()
		if errCode := VerifyAndSendTx(&txn); errCode != ErrNoError {
			return IPTRpc(errCode.Error())
		}
		return IPTRpc(BytesToHexString(txnHash.ToArrayReverse()))
	} else {
		var buffer bytes.Buffer
		txn.Serialize(&buffer)
		return IPTRpc(BytesToHexString(buffer.Bytes()))
	}
}

func createMultisigTransaction(params []interface{}) map[string]interface{} {
	if len(params) < 4 {
		return IPTRpcNil
	}
	var asset, from, address, value string
	switch params[0].(type) {
	case string:
		asset = params[0].(string)
	default:
		return IPTRpcInvalidParameter
	}
	switch params[1].(type) {
	case string:
		from = params[1].(string)
	default:
		return IPTRpcInvalidParameter
	}
	switch params[2].(type) {
	case string:
		address = params[2].(string)
	default:
		return IPTRpcInvalidParameter
	}
	switch params[3].(type) {
	case string:
		value = params[3].(string)
	default:
		return IPTRpcInvalidParameter
	}
	if Wallet == nil {
		return IPTRpc("error : wallet is not opened")
	}

	batchOut := sdk.BatchOut{
		Address: address,
		Value:   value,
	}
	tmp, err := HexStringToBytesReverse(asset)
	if err != nil {
		return IPTRpc("error: invalid asset ID")
	}
	var assetID Uint256
	if err := assetID.Deserialize(bytes.NewReader(tmp)); err != nil {
		return IPTRpc("error: invalid asset hash")
	}
	txn, err := sdk.MakeMultisigTransferTransaction(Wallet, assetID, from, batchOut)
	if err != nil {
		return IPTRpc("error: " + err.Error())
	}

	_, needsig, err := txn.ParseTransactionSig()
	if err != nil {
		return IPTRpc("error: " + err.Error())
	}
	if needsig == 0 {
		txnHash := txn.Hash()
		if errCode := VerifyAndSendTx(txn); errCode != ErrNoError {
			return IPTRpc(errCode.Error())
		}
		return IPTRpc(BytesToHexString(txnHash.ToArrayReverse()))
	} else {
		var buffer bytes.Buffer
		txn.Serialize(&buffer)
		return IPTRpc(BytesToHexString(buffer.Bytes()))
	}
}

func getBalance(params []interface{}) map[string]interface{} {
	if Wallet == nil {
		return IPTRpc("open wallet first")
	}
	type AssetInfo struct {
		AssetID string
		Value   string
	}
	balances := make(map[string][]*AssetInfo)
	accounts := Wallet.GetAccounts()
	coins := Wallet.GetCoins()
	for _, account := range accounts {
		assetList := []*AssetInfo{}
		programHash := account.ProgramHash
		for _, coin := range coins {
			if programHash == coin.Output.ProgramHash {
				var existed bool
				assetString := BytesToHexString(coin.Output.AssetID.ToArray())
				for _, info := range assetList {
					if info.AssetID == assetString {
						info.Value += coin.Output.Value.String()
						existed = true
						break
					}
				}
				if !existed {
					assetList = append(assetList, &AssetInfo{AssetID: assetString, Value: coin.Output.Value.String()})
				}
			}
		}
		address, _ := programHash.ToAddress()
		balances[address] = assetList
	}

	return IPTRpc(balances)
}
