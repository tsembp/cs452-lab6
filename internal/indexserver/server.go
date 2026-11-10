package indexserver

import (
	"sync"

	"github.com/ucy-coast/websearch/internal/invertedindex"
)

// IndexServer exposes search functionality over RPC.
type IndexServer struct {
	memoryMu  sync.Mutex
	index     *invertedindex.InvertedIndex
	useMemory bool
}

func NewIndexServer(index *invertedindex.InvertedIndex, useMemory bool) *IndexServer {
	s := &IndexServer{
		index:     index,
		useMemory: useMemory,
	}
	return s
}
