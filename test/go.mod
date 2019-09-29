module github.com/philippgille/gokv/test

go 1.12

require (
	github.com/go-test/deep v1.0.4
	github.com/philippgille/gokv v0.5.1-0.20190928144926-0fa470d8f0d4
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
