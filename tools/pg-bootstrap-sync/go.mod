module pg-bootstrap-sync

go 1.24.3

require (
	github.com/jackc/pglogrepl v0.0.0-20250509230407-a9884f6bd75a
	github.com/pganalyze/pg_query_go/v6 v6.1.0
	github.com/spf13/cobra v1.8.1
	kasho/pkg/kvbuffer v0.0.0-00010101000000-000000000000
	kasho/proto v0.0.0-00010101000000-000000000000
)

replace kasho/pkg/kvbuffer => ../../pkg/kvbuffer

replace kasho/proto => ../../proto/kasho/proto
