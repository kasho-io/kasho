module kasho/pkg/license

go 1.24.3

require (
	github.com/golang-jwt/jwt/v5 v5.2.0
	google.golang.org/grpc v1.72.1
	kasho/pkg/version v0.0.0
	kasho/proto v0.0.0
)

require (
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace (
	kasho/pkg/version => ../version
	kasho/proto => ../../proto/kasho/proto
)
