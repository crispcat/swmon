module swmon_check_stp

go 1.19

require (
	github.com/gosnmp/gosnmp v1.35.0
	swmon_shared v0.0.0-00010101000000-000000000000
)

require (
	github.com/hallidave/mibtool v0.2.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace swmon_shared => ../../shared
