package msg

import (
	. "IPT/common"
	"IPT/core/ledger"
	"IPT/core/transaction"
	"IPT/crypto"
	. "IPT/common/errors"
	"IPT/event"
	"IPT/msg/node"
	"IPT/msg/protocol"
)

type msger interface {
	GetTxnPool(byCount bool) map[Uint256]*transaction.Transaction
	Xmit(interface{}) error
	GetEvent(eventName string) *events.Event
	GetBookKeepersAddrs() ([]*crypto.PubKey, uint64)
	CleanSubmittedTransactions(block *ledger.Block) error
	GetNeighborNoder() []protocol.Noder
	Tx(buf []byte)
	AppendTxnPool(*transaction.Transaction, bool) ErrCode
}

func StartProtocol(pubKey *crypto.PubKey) protocol.Noder {
	msg := node.InitNode(pubKey)
	msg.ConnectSeeds()

	return msg
}
