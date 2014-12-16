package lochness

import (
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

// Context carries around data/structs needed for operations
type Context struct {
	etcd *etcd.Client
}

// NewContext creates a new context
func NewContext(e *etcd.Client) *Context {
	return &Context{
		etcd: e,
	}
}

// IsKeyNotFound is a helper to determine if the error is a key not found error
func IsKeyNotFound(err error) bool {
	return strings.Contains(err.Error(), "Key not found")
}
