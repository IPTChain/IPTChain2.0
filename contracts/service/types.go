package service

import (
	. "IPT/common"
	"IPT/core/asset"
	"IPT/core/ledger"
	"IPT/core/transaction"
	"IPT/msg/restful/common"
	. "IPT/msg/rpc"
	"IPT/contracts/states"
)

type AccountInfo struct {
	ProgramHash string
	IsFrozen    bool
	Balances    map[string]Fixed64
}

type AssetInfo struct {
	Name       string
	Precision  byte
	AssetType  byte
	RecordType byte
}

func GetHeaderInfo(header *ledger.Header) *BlockHead {
	h := header.Blockdata.Hash()
	return &BlockHead{
		Version:          header.Blockdata.Version,
		PrevBlockHash:    BytesToHexString(header.Blockdata.PrevBlockHash.ToArrayReverse()),
		TransactionsRoot: BytesToHexString(header.Blockdata.TransactionsRoot.ToArrayReverse()),
		Timestamp:        header.Blockdata.Timestamp,
		Height:           header.Blockdata.Height,
		ConsensusData:    header.Blockdata.ConsensusData,
		NextBookKeeper:   BytesToHexString(header.Blockdata.NextBookKeeper.ToArrayReverse()),
		Program: ProgramInfo{
			Code:      BytesToHexString(header.Blockdata.Program.Code),
			Parameter: BytesToHexString(header.Blockdata.Program.Parameter),
		},
		Hash: BytesToHexString(h.ToArrayReverse()),
	}

}

func GetBlockInfo(block *ledger.Block) *BlockInfo {
	blockInfo := common.GetBlockInfo(block)
	return &blockInfo
}

func GetTransactionInfo(transaction *transaction.Transaction) *Transactions {
	return TransArryByteToHexString(transaction)
}

func GetTransactionInputs(inputs []*transaction.UTXOTxInput) []UTXOTxInputInfo {
	inputList := make([]UTXOTxInputInfo, len(inputs))
	for k, v := range inputs {
		inputList[k].ReferTxID = BytesToHexString(v.ReferTxID.ToArrayReverse())
		inputList[k].ReferTxOutputIndex = v.ReferTxOutputIndex
	}
	return inputList
}

func GetTransactionOutputs(outputs []*transaction.TxOutput) []TxoutputInfo {
	outputList := make([]TxoutputInfo, len(outputs))
	for k, v := range outputs {
		outputList[k].AssetID = BytesToHexString(v.AssetID.ToArrayReverse())
		outputList[k].Value = v.Value.String()
		outputList[k].Address = BytesToHexString(v.ProgramHash.ToArrayReverse())
	}
	return outputList
}

func GetTransactionAttributes(attributes []*transaction.TxAttribute) []TxAttributeInfo {
	attributeList := make([]TxAttributeInfo, len(attributes))
	for k, v := range attributes {
		attributeList[k].Usage = v.Usage
		attributeList[k].Data = BytesToHexString(v.Data)
	}
	return attributeList
}

func GetAccountInfo(account *states.AccountState) *AccountInfo {
	balances := make(map[string]Fixed64)
	for k, v := range account.Balances {
		assetId := BytesToHexString(k.ToArrayReverse())
		balances[assetId] = v
	}
	return &AccountInfo{
		ProgramHash: BytesToHexString(account.ProgramHash.ToArrayReverse()),
		IsFrozen:    account.IsFrozen,
		Balances:    balances,
	}
}

func GetAssetInfo(asset *asset.Asset) *AssetInfo {
	return &AssetInfo{
		Name:       asset.Name,
		Precision:  asset.Precision,
		AssetType:  byte(asset.AssetType),
		RecordType: byte(asset.RecordType),
	}
}
