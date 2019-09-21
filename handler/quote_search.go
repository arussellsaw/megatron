package handler

import (
	"sort"

	"github.com/monzo/slog"
	"github.com/monzo/terrors"
	"github.com/monzo/typhon"

	"github.com/arussellsaw/megatron/domain/library"
)

type QuoteSearchRequest struct {
	Query string `form:"q" json:"query"`
}

type QuoteSearchResponse struct {
	Results []library.SearchResult `json:"results"`
}

func handleQuoteSearch(req typhon.Request) typhon.Response {
	body := QuoteSearchRequest{}
	if err := req.Decode(&body); err != nil {
		return typhon.Response{Error: terrors.Wrap(err, nil)}
	}
	params := map[string]string{
		"query": body.Query,
	}

	slog.Info(req, "Handling search for %s", body.Query, params)

	res, err := DefaultIndex.Search(body.Query)
	if err != nil {
		return typhon.Response{Error: terrors.Wrap(err, params)}
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Confidence < res[j].Confidence
	})

	if len(res) > 3 {
		res = res[3:]
	}

	return req.Response(QuoteSearchResponse{res})
}
