package webserver

import (
	"encoding/json"
	"fmt"
	"github.com/ucy-coast/websearch/internal/rpc_api"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sort"
	"time"
)

// RootHandler serves the static HTML file
func (s *WebServer) RootHandler(w http.ResponseWriter, r *http.Request) {
	htmlContent, err := os.ReadFile(s.htmlPath)
	if err != nil {
		http.Error(w, "Unable to load HTML file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(htmlContent)
}

// SearchHandler queries all shard RPC servers concurrently and merges the results
func (s *WebServer) SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing query parameter", http.StatusBadRequest)
		return
	}

	results, err := s.searchAcrossShards(query, s.topK)
	if err != nil {
		http.Error(w, "search failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// searchAcrossShards sends the search request to all shard servers and aggregates the results.
func (s *WebServer) searchAcrossShards(query string, topK int) ([]rpc_api.SearchResult, error) {

	results := make([]rpc_api.SearchResult, 0)
	if topK <= 0 {
		return results, nil
	}
	type shardResponse struct {
		results []rpc_api.SearchResult
		err     error
	}
	responses := make(chan shardResponse, len(s.shardRPCAddrs))
	for _, addr := range s.shardRPCAddrs {
		go func(addr string) {
			conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
			if err != nil {
				responses <- shardResponse{err: fmt.Errorf("connect to shard %s: %w", addr, err)}
				return
			}
			defer conn.Close()
			if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
				responses <- shardResponse{err: err}
				return
			}
			client := rpc.NewClient(conn)
			defer client.Close()
			// Fetch all candidates: a document can collect matches from several shards.
			args := rpc_api.SearchArgs{Query: query, TopK: int(^uint(0) >> 1)}
			var reply []rpc_api.SearchResult
			err = client.Call("IndexServer.Search", &args, &reply)
			if err != nil {
				err = fmt.Errorf("search shard %s: %w", addr, err)
			}
			responses <- shardResponse{results: reply, err: err}
		}(addr)
	}
	matchesByDoc := make(map[string]map[string]struct{})
	for range s.shardRPCAddrs {
		response := <-responses
		if response.err != nil {
			return nil, response.err
		}
		for _, result := range response.results {
			if matchesByDoc[result.Document] == nil {
				matchesByDoc[result.Document] = make(map[string]struct{})
			}
			for _, keyword := range result.Matches {
				matchesByDoc[result.Document][keyword] = struct{}{}
			}
		}
	}
	for doc, keywords := range matchesByDoc {
		matches := make([]string, 0, len(keywords))
		for keyword := range keywords {
			matches = append(matches, keyword)
		}
		sort.Strings(matches)
		results = append(results, rpc_api.SearchResult{Document: doc, Matches: matches})
	}
	sort.Slice(results, func(i, j int) bool {
		if len(results[i].Matches) != len(results[j].Matches) {
			return len(results[i].Matches) > len(results[j].Matches)
		}
		return results[i].Document < results[j].Document
	})
	if len(results) > topK {
		results = results[:topK]
	}
	return results, nil

}
