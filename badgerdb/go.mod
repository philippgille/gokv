module github.com/philippgille/gokv/badgerdb

go 1.12

require (
	github.com/AndreasBriese/bbloom v0.0.0-20190825152654-46b345b51c96 // indirect
	github.com/dgraph-io/badger v1.6.0
	github.com/go-test/deep v1.0.4 // indirect
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/philippgille/gokv v0.5.0
	github.com/stretchr/testify v1.4.0 // indirect
	golang.org/x/net v0.0.0-20190926025831-c00fd9afed17 // indirect
	golang.org/x/sys v0.0.0-20190927073244-c990c680b611 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
