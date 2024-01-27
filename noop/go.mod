module github.com/philippgille/gokv/noop

go 1.20

require (
	github.com/philippgille/gokv v0.6.0
	github.com/philippgille/gokv/util v0.6.0
)

replace (
	github.com/philippgille/gokv => ../
	github.com/philippgille/gokv/encoding => ../encoding
	github.com/philippgille/gokv/sql => ../sql
	github.com/philippgille/gokv/test => ../test
	github.com/philippgille/gokv/util => ../util
)
