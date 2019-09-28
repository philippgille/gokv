module github.com/philippgille/gokv/ignite

go 1.12

require (
	github.com/amsokol/ignite-go-client v0.12.2
	github.com/go-test/deep v1.0.4 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/philippgille/gokv v0.0.0-00010101000000-000000000000
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
