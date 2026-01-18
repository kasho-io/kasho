module generate

go 1.24.3

require (
	github.com/brianvoe/gofakeit/v7 v7.0.2
	github.com/google/uuid v1.6.0
	golang.org/x/crypto v0.39.0
	kasho/pkg/dialect v0.0.0
)

require (
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/grpc v1.72.1 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	kasho/proto v0.0.0-00010101000000-000000000000 // indirect
)

replace kasho/pkg/dialect => ../../../pkg/dialect

replace kasho/proto => ../../../proto/kasho/proto
