package rpc

import (
	. "IPT/common/config"
	"IPT/common/log"
	"net/http"
	"strconv"
)

const (
	//	LocalHost = "127.0.0.1"

	LocalHost = "0.0.0.0"
)

func StartRPCServer() {
	log.Debug()
	http.HandleFunc("/", Handle)

	HandleFunc("getbestblockhash", getBestBlockHash)
	HandleFunc("getblock", getBlock)
	HandleFunc("getblockcount", getBlockCount)
	HandleFunc("getblockhash", getBlockHash)
	HandleFunc("getconnectioncount", getConnectionCount)
	HandleFunc("getrawmempool", getRawMemPool)
	HandleFunc("getrawtransaction", getRawTransaction)
	HandleFunc("sendrawtransaction", sendRawTransaction)
	HandleFunc("getversion", getVersion)
	HandleFunc("getneighbor", getNeighbor)
	HandleFunc("getnodestate", getNodeState)

	HandleFunc("setdebuginfo", setDebugInfo)
	HandleFunc("lockasset", lockAsset)
	HandleFunc("createmultisigtransaction", createMultisigTransaction)
	HandleFunc("signmultisigtransaction", signMultisigTransaction)
	HandleFunc("addaccount", addAccount)
	HandleFunc("openwallet", openWallet)
	HandleFunc("closewallet", closeWallet)
	HandleFunc("sendtoaddress", sendToAddress)
	HandleFunc("createAccountForCust", createAccountForCust)

	err := http.ListenAndServe(LocalHost+":"+strconv.Itoa(Parameters.HttpJsonPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
