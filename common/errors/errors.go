package errors

import (
	"errors"
)

const callStackDepth = 10

type DetailError interface {
	error
	ErrCoder
	CallStacker
	GetRoot()  error
}


func  NewErr(errmsg string) error {
	return errors.New(errmsg)
}

func NewDetailErr(err error,errcode ErrCode,errmsg string) DetailError{
	if err == nil {return nil}

	IPTerr, ok := err.(IPTError)
	if !ok {
		IPTerr.root = err
		IPTerr.errmsg = err.Error()
		IPTerr.callstack = getCallStack(0, callStackDepth)
		IPTerr.code = errcode

	}
	if errmsg != "" {
		IPTerr.errmsg = errmsg + ": " + IPTerr.errmsg
	}


	return IPTerr
}

func RootErr(err error) error {
	if err, ok := err.(DetailError); ok {
		return err.GetRoot()
	}
	return err
}



