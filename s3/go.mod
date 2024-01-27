module github.com/philippgille/gokv/s3

go 1.20

require (
	github.com/aws/aws-sdk-go v1.49.16
	github.com/philippgille/gokv/encoding v0.6.0
	github.com/philippgille/gokv/test v0.6.0
	github.com/philippgille/gokv/util v0.6.0
)

require (
	github.com/go-test/deep v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/philippgille/gokv v0.6.0 // indirect
)

replace (
	github.com/philippgille/gokv => ../
	github.com/philippgille/gokv/encoding => ../encoding
	github.com/philippgille/gokv/sql => ../sql
	github.com/philippgille/gokv/test => ../test
	github.com/philippgille/gokv/util => ../util
)
