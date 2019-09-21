package handler

import (
	"github.com/monzo/typhon"

	"github.com/arussellsaw/megatron/domain/library"
)

// DefaultIndex is the default index used by search handlers
var DefaultIndex *library.Index

// Router returns a http.Handler with all handler routes configured
func Router() typhon.Service {
	r := typhon.Router{}

	r.Register("GET", "/api/search", handleQuoteSearch)
	r.Register("GET", "/api/render", handleRenderGIF)

	return r.Serve()
}
