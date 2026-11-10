package indexserver

import (
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/ucy-coast/websearch/internal/invertedindex"
	"github.com/ucy-coast/websearch/internal/rpc_api"
)

func TestSearchModes(t *testing.T) {
	for _, memory := range []bool{false, true} {
		name := "disk"
		if memory {
			name = "memory"
		}
		t.Run(name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "index.txt")
			writeLines(t, file, []string{"apple: 2 doc1,doc2", "banana: 1 doc1"})
			idx, err := invertedindex.NewInvertedIndex([]string{file})
			if err != nil {
				t.Fatal(err)
			}
			server := NewIndexServer(idx, memory)
			// Exercise simultaneous first requests as well as normalization and top-K.
			var wg sync.WaitGroup
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					var reply []rpc_api.SearchResult
					err := server.Search(&rpc_api.SearchArgs{Query: " APPLE\tBanana apple ", TopK: 1}, &reply)
					want := []rpc_api.SearchResult{{Document: "doc1", Matches: []string{"apple", "banana"}}}
					if err != nil || !reflect.DeepEqual(reply, want) {
						t.Errorf("Search = %v, %v; want %v", reply, err, want)
					}
				}()
			}
			wg.Wait()
			for _, args := range []rpc_api.SearchArgs{{Query: "missing", TopK: 10}, {Query: "apple", TopK: 0}, {Query: "apple", TopK: -1}} {
				var reply []rpc_api.SearchResult
				if err := server.Search(&args, &reply); err != nil || len(reply) != 0 {
					t.Errorf("Search(%+v) = %v, %v", args, reply, err)
				}
			}
		})
	}
}
