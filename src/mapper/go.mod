module swmon_mapper

go 1.19

require (
	github.com/go-ping/ping v1.1.0
	github.com/gosnmp/gosnmp v1.35.0
	gopkg.in/yaml.v3 v3.0.1
	swmon_shared v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/uuid v1.2.0 // indirect
	github.com/hallidave/mibtool v0.2.0 // indirect
	golang.org/x/net v0.0.0-20210610124326-52da8fb2a613 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210423082822-04245dca01da // indirect
)

replace swmon_shared => ../shared
