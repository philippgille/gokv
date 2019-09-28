module github.com/philippgille/gokv/bbolt

go 1.12

require (
	github.com/go-test/deep v1.0.4 // indirect
	github.com/philippgille/gokv v0.5.0
	go.etcd.io/bbolt v1.3.3
	golang.org/x/sys v0.0.0-20190927073244-c990c680b611 // indirect
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
