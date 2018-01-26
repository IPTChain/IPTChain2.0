package errors

type IPTError struct {
	errmsg string
	callstack *CallStack
	root error
	code ErrCode
}

func (e IPTError) Error() string {
	return e.errmsg
}

func (e IPTError) GetErrCode()  ErrCode {
	return e.code
}

func (e IPTError) GetRoot()  error {
	return e.root
}

func (e IPTError) GetCallStack()  *CallStack {
	return e.callstack
}
