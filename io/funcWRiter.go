package io

type FuncWriter struct {
	Callback func (p []byte) (n int, err error)
}

func (fw FuncWriter) Write(p []byte) (n int, err error) {
	return fw.Callback(p)
}
