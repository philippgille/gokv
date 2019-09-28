module github.com/philippgille/gokv/memcached

go 1.12

require (
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b
	github.com/go-test/deep v1.0.4 // indirect
	github.com/philippgille/gokv v0.5.0
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
