module github.com/philippgille/gokv/examples

go 1.12

require (
	github.com/philippgille/gokv v0.5.1-0.20190929161952-f31a8dbcad2a
	github.com/philippgille/gokv/redis v0.5.1-0.20190929161952-f31a8dbcad2a
)

replace github.com/philippgille/gokv/redis => ../redis
