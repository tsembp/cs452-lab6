package indexserver

import (
	"fmt"
	"strings"

	"github.com/ucy-coast/websearch/internal/invertedindex"
	"github.com/ucy-coast/websearch/internal/rpc_api"
)

func (s *IndexServer) Search(args *rpc_api.SearchArgs, reply *[]rpc_api.SearchResult) error {
	queryWords := strings.Fields(strings.ToLower(args.Query))

	var results []invertedindex.Result
	var err error
	if s.useMemory {
		// The index loads lazily, so concurrent RPCs must not initialize it together.
		s.memoryMu.Lock()
		results, err = s.index.SearchInMemoryTopK(queryWords, args.TopK)
		s.memoryMu.Unlock()
	} else {
		results, err = s.index.SearchScanFiles(queryWords)
		if err == nil {
			k := args.TopK
			if k < 0 {
				k = 0
			}
			if k < len(results) {
				results = results[:k]
			}
		}
	}
	if err != nil {
		return fmt.Errorf("search index: %w", err)
	}

	*reply = make([]rpc_api.SearchResult, len(results))
	for i, result := range results {
		(*reply)[i] = rpc_api.SearchResult{
			Document: result.Document,
			Matches:  result.Matches,
		}
	}
	return nil
}
