package errors

import "errors"

var (
	ErrAssetNameInvalid      = errors.New("asset name invalid")
	ErrAssetAmountInvalid    = errors.New("asset amount invalid")
	ErrAssetPrecisionInvalid = errors.New("asset precision invalid")
)
