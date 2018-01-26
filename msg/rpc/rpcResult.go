package rpc

var (
	IPTRpcInvalidHash        = responsePacking("invalid hash")
	IPTRpcInvalidBlock       = responsePacking("invalid block")
	IPTRpcInvalidTransaction = responsePacking("invalid transaction")
	IPTRpcInvalidParameter   = responsePacking("invalid parameter")

	IPTRpcUnknownBlock       = responsePacking("unknown block")
	IPTRpcUnknownTransaction = responsePacking("unknown transaction")

	IPTRpcNil           = responsePacking(nil)
	IPTRpcUnsupported   = responsePacking("Unsupported")
	IPTRpcInternalError = responsePacking("internal error")
	IPTRpcIOError       = responsePacking("internal IO error")
	IPTRpcAPIError      = responsePacking("internal API error")
	IPTRpcSuccess       = responsePacking(true)
	IPTRpcFailed        = responsePacking(false)

	// error code for wallet
	IPTRpcWalletAlreadyExists = responsePacking("wallet already exist")
	IPTRpcWalletNotExists     = responsePacking("wallet doesn't exist")

	IPTRpc = responsePacking
)
