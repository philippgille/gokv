module github.com/philippgille/gokv/examples

go 1.12

require (
	github.com/philippgille/gokv v0.5.0
	github.com/philippgille/gokv/redis v0.5.0
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../

replace github.com/philippgille/gokv/redis => ../redis
