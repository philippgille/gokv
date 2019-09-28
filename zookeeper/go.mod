module github.com/philippgille/gokv/zookeeper

go 1.12

require (
	github.com/go-test/deep v1.0.4 // indirect
	github.com/philippgille/gokv v0.5.0
	github.com/samuel/go-zookeeper v0.0.0-20190923202752-2cc03de413da
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
