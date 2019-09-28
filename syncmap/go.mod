module github.com/philippgille/gokv/syncmap

go 1.12

require (
	github.com/go-test/deep v1.0.4 // indirect
	github.com/philippgille/gokv v0.5.0
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
