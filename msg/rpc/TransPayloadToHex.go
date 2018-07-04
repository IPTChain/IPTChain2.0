package rpc

import (
	. "IPT/common"
	"IPT/core/asset"
	. "IPT/core/transaction"
	"IPT/core/transaction/payload"
	"bytes"
)

type PayloadInfo interface{}

//implement PayloadInfo define BookKeepingInfo
type BookKeepingInfo struct {
	Nonce uint64
}

//implement PayloadInfo define DeployCodeInfo
type FunctionCodeInfo struct {
	Code           string
	ParameterTypes []int
	ReturnType     int
	CodeHash       string
}

type DeployCodeInfo struct {
	Code        *FunctionCodeInfo
	Name        string
	Version     string
	Author      string
	Email       string
	Description string
	Language    int
	ProgramHash string
}

type IssuerInfo struct {
	X, Y string
}

//implement PayloadInfo define RegisterAssetInfo
type RegisterAssetInfo struct {
	Asset      *asset.Asset
	Amount     string
	Issuer     IssuerInfo
	Controller string
}

type LockAssetInfo struct {
	Address    string
	AssetID    string
	Amount     string
	LockHeight uint32
}

type RecordInfo struct {
	RecordType string
	RecordData string
}

type BookkeeperInfo struct {
	PubKey     string
	Action     string
	Issuer     IssuerInfo
	Controller string
}

type DataFileInfo struct {
	IPFSPath string
	Filename string
	Note     string
	Issuer   IssuerInfo
}

type PrivacyPayloadInfo struct {
	PayloadType uint8
	Payload     string
	EncryptType uint8
	EncryptAttr string
}

func TransPayloadToHex(p Payload) PayloadInfo {
	switch object := p.(type) {
	case *payload.BookKeeping:
		obj := new(BookKeepingInfo)
		obj.Nonce = object.Nonce
		return obj
	case *payload.BookKeeper:
		obj := new(BookkeeperInfo)
		encodedPubKey, _ := object.PubKey.EncodePoint(true)
		obj.PubKey = BytesToHexString(encodedPubKey)
		if object.Action == payload.BookKeeperAction_ADD {
			obj.Action = "add"
		} else if object.Action == payload.BookKeeperAction_SUB {
			obj.Action = "sub"
		} else {
			obj.Action = "nil"
		}
		obj.Issuer.X = object.Issuer.X.String()
		obj.Issuer.Y = object.Issuer.Y.String()

		return obj
	case *payload.IssueAsset:
	case *payload.TransferAsset:
	case *payload.DeployCode:
		obj := new(DeployCodeInfo)
		obj.Code = new(FunctionCodeInfo)
		obj.Code.Code = BytesToHexString(object.Code.Code)
		var params []int
		for _, v := range object.Code.ParameterTypes {
			params = append(params, int(v))
		}
		obj.Code.ParameterTypes = params
		obj.Code.ReturnType = int(object.Code.ReturnType)
		codeHash := object.Code.CodeHash()
		obj.Code.CodeHash = BytesToHexString(codeHash.ToArrayReverse())
		obj.Name = object.Name
		obj.Version = object.CodeVersion
		obj.Author = object.Author
		obj.Email = object.Email
		obj.Description = object.Description
		obj.Language = int(object.Language)
		obj.ProgramHash = BytesToHexString(object.ProgramHash.ToArrayReverse())
		return obj
	case *payload.RegisterAsset:
		obj := new(RegisterAssetInfo)
		obj.Asset = object.Asset
		obj.Amount = object.Amount.String()
		obj.Issuer.X = object.Issuer.X.String()
		obj.Issuer.Y = object.Issuer.Y.String()
		obj.Controller = BytesToHexString(object.Controller.ToArrayReverse())
		return obj
	case *payload.LockAsset:
		obj := new(LockAssetInfo)
		address, _ := object.ProgramHash.ToAddress()
		obj.Address = address
		obj.AssetID = BytesToHexString(object.AssetID.ToArrayReverse())
		obj.Amount = object.Amount.String()
		obj.LockHeight = object.UnlockHeight
		return obj
	case *payload.Record:
		obj := new(RecordInfo)
		obj.RecordType = object.RecordType
		obj.RecordData = BytesToHexString(object.RecordData)
		return obj
	case *payload.PrivacyPayload:
		obj := new(PrivacyPayloadInfo)
		obj.PayloadType = uint8(object.PayloadType)
		obj.Payload = BytesToHexString(object.Payload)
		obj.EncryptType = uint8(object.EncryptType)
		bytesBuffer := bytes.NewBuffer([]byte{})
		object.EncryptAttr.Serialize(bytesBuffer)
		obj.EncryptAttr = BytesToHexString(bytesBuffer.Bytes())
		return obj
	case *payload.DataFile:
		obj := new(DataFileInfo)
		obj.IPFSPath = object.IPFSPath
		obj.Filename = object.Filename
		obj.Note = object.Note
		obj.Issuer.X = object.Issuer.X.String()
		obj.Issuer.Y = object.Issuer.Y.String()
		return obj
	}
	return nil
}
