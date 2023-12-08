module swmon_mapper

go 1.18

require (
	github.com/go-ping/ping v1.1.0
	gopkg.in/yaml.v3 v3.0.1
	swmon_shared v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/uuid v1.4.0 // indirect
	github.com/gosnmp/gosnmp v1.37.0 // indirect
	github.com/hallidave/mibtool v0.2.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
)

replace swmon_shared => ../shared
