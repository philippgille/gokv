module github.com/philippgille/gokv/examples

go 1.12

require (
	github.com/philippgille/gokv v0.5.1-0.20190929161952-f31a8dbcad2a
	github.com/philippgille/gokv/redis v0.0.0-20191001201555-5ac9a20de634
)

replace github.com/philippgille/gokv/redis => ../redis
