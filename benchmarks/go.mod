module github.com/Baselyne-Systems/bulkhead/benchmarks

go 1.24.0

require (
	github.com/Baselyne-Systems/bulkhead/control-plane v0.0.0
	google.golang.org/grpc v1.78.0
)

require (
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260128011058-8636f8732409 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/Baselyne-Systems/bulkhead/control-plane => ../control-plane
