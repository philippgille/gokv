module github.com/philippgille/gokv/examples

go 1.13

require (
	github.com/philippgille/gokv v0.5.1-0.20191011213304-eb77f15b9c61
	github.com/philippgille/gokv/redis v0.0.0
)

replace github.com/philippgille/gokv/redis => ../redis
