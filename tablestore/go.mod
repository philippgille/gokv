module github.com/philippgille/gokv/tablestore

go 1.12

require (
	github.com/aliyun/aliyun-tablestore-go-sdk v4.1.3+incompatible
	github.com/go-test/deep v1.0.4 // indirect
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/philippgille/gokv v0.5.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
