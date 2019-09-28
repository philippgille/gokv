module github.com/philippgille/gokv/mongodb

go 1.12

require (
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/go-test/deep v1.0.4 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/philippgille/gokv v0.5.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
