# CS452 Lab 6 — Sharding

A distributed Go search system with an HTTP frontend and multiple backend index shards communicating through `net/rpc`.

## Run

From the repository root, build both servers:

```bash
make indexserver webserver
```

Start each backend in a separate terminal:

```bash
./bin/indexserver -rpc_addr 127.0.0.1:9090 -index_files test/data/invertedindex-medium-0.txt
```

```bash
./bin/indexserver -rpc_addr 127.0.0.1:9091 -index_files test/data/invertedindex-medium-1.txt
```

Then start the frontend in another terminal:

```bash
./bin/webserver -addr 127.0.0.1:8080 -shards 127.0.0.1:9090,127.0.0.1:9091 -topk 10
```

Query the API from the host:

```bash
curl http://127.0.0.1:8080/api/search?q=adventure%20grief
```

The web interface is available on port 8080; forward that port in VS Code when running on CloudLab. Add `-use_memory` to a backend command to use its lazily loaded in-memory index instead of scanning files.

## Implemented

- Backend RPC search with lowercase query normalization and top-K limiting.
- Concurrent frontend queries to all shards using goroutines and a channel.
- Document merging with unique matched keywords, ranked by match count and then document name.
- Configurable global top-K, connection/RPC timeouts, and shard error handling.
- Safe concurrent access to the lazily loaded in-memory index.

The frontend fetches all candidates before applying global top-K, so documents can accumulate matches across shards.

## Tests

```bash
go test -race ./...
```

All tests pass, including RPC merging, ranking, duplicate keywords, unavailable shards, and both backend search modes. The provided one-shard and two-shard scripts were also exercised successfully; both exit with status 143 during their server shutdown step.
