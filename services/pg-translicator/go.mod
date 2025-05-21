module pg-translicator

go 1.24.3

require (
	github.com/brianvoe/gofakeit/v7 v7.0.2
	google.golang.org/grpc v1.64.0
	gopkg.in/yaml.v3 v3.0.1
	pg-change-stream v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

replace pg-change-stream => ../pg-change-stream
