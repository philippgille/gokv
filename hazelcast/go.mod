module github.com/philippgille/gokv/hazelcast

go 1.12

require (
	github.com/apache/thrift v0.12.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-test/deep v1.0.4 // indirect
	github.com/hazelcast/hazelcast-go-client v0.0.0-20190530123621-6cf767c2f31a
	github.com/kr/pretty v0.1.0 // indirect
	github.com/philippgille/gokv v0.5.0
	github.com/stretchr/testify v1.4.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
