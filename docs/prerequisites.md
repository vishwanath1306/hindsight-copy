# Pre-requisites

* gcc
* golang 1.11 or higher (to support go modules)

# Environment variables
* You may need to add `{hindsight_dir}/agent` to your `$GOPATH`
* Hindsight requires the following variables for using cgo:
```
export CGO_LDFLAGS_ALLOW=".*"
```
* (Unsure) maybe increase gomaxprocs:
```
export GOMAXPROCS=10
```

# Recompiling protocols

If during development you need to recompile the protocol buffer definitions, you will need a protocol buffers compiler plus the grpc golang extensions.