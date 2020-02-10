module github.com/philippgille/gokv/pogreb

go 1.13

require (
	github.com/akrylysov/pogreb v0.8.3
	github.com/philippgille/gokv v0.6.0
	github.com/philippgille/gokv/encoding v0.6.0
	github.com/philippgille/gokv/test v0.6.0
	github.com/philippgille/gokv/util v0.6.0
)

replace github.com/philippgille/gokv/util v0.6.0 => ../util
