module github.com/philippgille/gokv/postgresql

go 1.12

require (
	github.com/go-test/deep v1.0.4 // indirect
	github.com/lib/pq v1.2.0
	github.com/philippgille/gokv v0.5.0
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
