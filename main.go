package main

import (
	"os"
	"runtime"
	"time"

	"IPT/account"
	"IPT/common/config"
	"IPT/common/log"
	"IPT/consensus/dbft"
	"IPT/core/ledger"
	"IPT/core/store/ChainStore"
	"IPT/core/transaction"
	"IPT/crypto"
	"IPT/msg"
	"IPT/msg/rpc"
	"IPT/msg/nodeinfo"
	"IPT/msg/restful"
	"IPT/msg/socket"
	"IPT/msg/protocol"
)

const (
	DefaultMultiCoreNum = 4
)

func init() {
	log.Init(log.Path, log.Stdout)
	var coreNum int
	if config.Parameters.MultiCoreNum > DefaultMultiCoreNum {
		coreNum = int(config.Parameters.MultiCoreNum)
	} else {
		coreNum = DefaultMultiCoreNum
	}
	log.Debug("The Core number is ", coreNum)
	runtime.GOMAXPROCS(coreNum)
}

func main() {
	var client account.Client
	var acct *account.Account
	var blockChain *ledger.Blockchain
	var err error
	var noder protocol.Noder
	log.Trace("Node version: ", config.Version)

	if len(config.Parameters.BookKeepers) < account.DefaultBookKeeperCount {
		log.Fatal("At least ", account.DefaultBookKeeperCount, " BookKeepers should be set at config.json")
		os.Exit(1)
	}

	log.Info("0. Loading the Ledger")
	ledger.DefaultLedger = new(ledger.Ledger)
	ledger.DefaultLedger.Store, err = ChainStore.NewLedgerStore()
	defer ledger.DefaultLedger.Store.Close()
	if err != nil {
		log.Fatal("open LedgerStore err:", err)
		os.Exit(1)
	}
	ledger.DefaultLedger.Store.InitLedgerStore(ledger.DefaultLedger)
	transaction.TxStore = ledger.DefaultLedger.Store
	crypto.SetAlg(config.Parameters.EncryptAlg)

	log.Info("1. BlockChain init")
	ledger.StandbyBookKeepers = account.GetBookKeepers()
	blockChain, err = ledger.NewBlockchainWithGenesisBlock(ledger.StandbyBookKeepers)
	if err != nil {
		log.Fatal(err, "  BlockChain generate failed")
		goto ERROR
	}
	ledger.DefaultLedger.Blockchain = blockChain

	if protocol.DATANODENAME == config.Parameters.NodeType {
		// Generate a new account for data node
		acct, err = account.NewAccount()
		if err != nil {
			log.Fatal("Can't create local account.")
			goto ERROR
		}
	} else {
		log.Info("2. Open the account")
		client = account.GetClient()
		if client == nil {
			log.Fatal("Can't get local account.")
			goto ERROR
		}
		acct, err = client.GetDefaultAccount()
		if err != nil {
			log.Fatal(err)
			goto ERROR
		}
		// expose wallet to httpjson interface
		httpjsonrpc.Wallet = client
	}
	log.Info("The Node's PublicKey ", acct.PublicKey)

	log.Info("3. Start the P2P networks")
	noder = net.StartProtocol(acct.PublicKey)
	httpjsonrpc.RegistRpcNode(noder)
	time.Sleep(10 * time.Second)
	noder.SyncNodeHeight()
	noder.WaitForFourPeersStart()
	noder.WaitForSyncBlkFinish()
	if protocol.VERIFYNODENAME == config.Parameters.NodeType {
		log.Info("4. Start DBFT Services")
		dbftServices := dbft.NewDbftService(client, "logdbft", noder)
		httpjsonrpc.RegistDbftService(dbftServices)
		go dbftServices.Start()
		time.Sleep(5 * time.Second)
	}

	log.Info("--Start the RPC interface")
	go httpjsonrpc.StartRPCServer()
	go httprestful.StartServer(noder)
	go httpwebsocket.StartServer(noder)
	if config.Parameters.HttpInfoStart {
		go httpnodeinfo.StartServer(noder)
	}

	for {
		time.Sleep(dbft.GenBlockTime)
		log.Trace("BlockHeight = ", ledger.DefaultLedger.Blockchain.BlockHeight)
		isNeedNewFile := log.CheckIfNeedNewFile()
		if isNeedNewFile == true {
			log.ClosePrintLog()
			log.Init(log.Path, os.Stdout)
		}
	}

ERROR:
	os.Exit(1)
}
