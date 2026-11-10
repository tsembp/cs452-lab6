package webserver

import (
	"github.com/ucy-coast/websearch/internal/rpc_api"
	"net"
	"net/rpc"
	"reflect"
	"testing"
)

type TestShard struct{ Results []rpc_api.SearchResult }

func (s *TestShard) Search(args *rpc_api.SearchArgs, reply *[]rpc_api.SearchResult) error {
	*reply = s.Results
	return nil
}
func startTestShard(t *testing.T, results []rpc_api.SearchResult) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	server := rpc.NewServer()
	if err := server.RegisterName("IndexServer", &TestShard{Results: results}); err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.ServeConn(conn)
		}
	}()
	return listener.Addr().String()
}
func TestSearchAcrossShardsMerge(t *testing.T) {
	a := startTestShard(t, []rpc_api.SearchResult{
		{Document: "shared", Matches: []string{"apple"}},
		{Document: "z", Matches: []string{"apple"}},
	})
	b := startTestShard(t, []rpc_api.SearchResult{
		{Document: "shared", Matches: []string{"banana", "apple"}},
		{Document: "a", Matches: []string{"banana"}},
	})
	server := NewWebServer("", 2, []string{a, b})
	got, err := server.searchAcrossShards("apple banana", server.topK)
	want := []rpc_api.SearchResult{
		{Document: "shared", Matches: []string{"apple", "banana"}},
		{Document: "a", Matches: []string{"banana"}},
	}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, %v; want %v", got, err, want)
	}
}
func TestSearchAcrossShardsFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()
	server := NewWebServer("", 10, []string{addr})
	if _, err := server.searchAcrossShards("apple", 10); err == nil {
		t.Fatal("expected unavailable shard error")
	}
}
