module github.com/philippgille/gokv/examples/protobuf_encoding

go 1.20

require (
	github.com/philippgille/gokv v0.6.0
	github.com/philippgille/gokv/encoding/protobuf v0.0.0
	github.com/philippgille/gokv/gomap v0.6.0
	google.golang.org/protobuf v1.32.0
)

replace github.com/philippgille/gokv/encoding/protobuf => ../../encoding/protobuf

require (
	github.com/philippgille/gokv/encoding v0.6.0 // indirect
	github.com/philippgille/gokv/util v0.6.0 // indirect
)
