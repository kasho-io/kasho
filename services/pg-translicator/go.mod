module pg-translicator

go 1.24.3

require (
	github.com/brianvoe/gofakeit/v7 v7.0.2
	github.com/lib/pq v1.10.9
	google.golang.org/grpc v1.72.1
	gopkg.in/yaml.v3 v3.0.1
	kasho/proto v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace pg-change-stream => ../pg-change-stream

replace kasho/proto => ../../proto/kasho/proto
