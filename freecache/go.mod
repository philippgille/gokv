module github.com/philippgille/gokv/freecache

go 1.12

require (
	github.com/OneOfOne/xxhash v1.2.5 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/coocood/freecache v1.1.0
	github.com/go-test/deep v1.0.4 // indirect
	github.com/philippgille/gokv v0.5.0
	github.com/spaolacci/murmur3 v1.1.0 // indirect
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
