module pg-change-stream

go 1.24.3

require (
	github.com/jackc/pglogrepl v0.0.0-20250509230407-a9884f6bd75a
	github.com/jackc/pgx/v5 v5.5.4
	github.com/lib/pq v1.10.9
	google.golang.org/grpc v1.72.1
	kasho/pkg/kvbuffer v0.0.0-00010101000000-000000000000
	kasho/pkg/license v0.0.0-00010101000000-000000000000
	kasho/pkg/types v0.0.0-00010101000000-000000000000
	kasho/proto v0.0.0-00010101000000-000000000000
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang-jwt/jwt/v5 v5.2.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/redis/go-redis/v9 v9.8.0 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	kasho/pkg/version v0.0.0 // indirect
)

replace kasho/pkg/kvbuffer => ../../pkg/kvbuffer

replace kasho/pkg/license => ../../pkg/license

replace kasho/pkg/types => ../../pkg/types

replace kasho/pkg/version => ../../pkg/version

replace kasho/proto => ../../proto/kasho/proto
