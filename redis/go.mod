module github.com/philippgille/gokv/redis

go 1.12

require (
	github.com/go-redis/redis v6.15.5+incompatible
	github.com/go-test/deep v1.0.4 // indirect
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/onsi/ginkgo v1.10.1 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/philippgille/gokv v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20190926025831-c00fd9afed17 // indirect
	golang.org/x/sys v0.0.0-20190927073244-c990c680b611 // indirect
	golang.org/x/text v0.3.2 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
