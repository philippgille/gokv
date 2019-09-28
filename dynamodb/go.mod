module github.com/philippgille/gokv/dynamodb

go 1.12

require (
	github.com/aws/aws-sdk-go v1.25.1
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-test/deep v1.0.4 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/philippgille/gokv v0.5.0
	github.com/stretchr/testify v1.4.0 // indirect
	golang.org/x/net v0.0.0-20190926025831-c00fd9afed17 // indirect
	golang.org/x/text v0.3.2 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
