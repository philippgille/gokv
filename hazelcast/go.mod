module github.com/philippgille/gokv/hazelcast

go 1.20

require (
	github.com/hazelcast/hazelcast-go-client v1.4.1
	github.com/philippgille/gokv/encoding v0.6.0
	github.com/philippgille/gokv/test v0.6.0
	github.com/philippgille/gokv/util v0.6.0
)

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/go-test/deep v1.1.0 // indirect
	github.com/philippgille/gokv v0.6.0 // indirect
	github.com/shirou/gopsutil/v3 v3.21.5 // indirect
	github.com/tklauser/go-sysconf v0.3.4 // indirect
	github.com/tklauser/numcpus v0.2.1 // indirect
	golang.org/x/sys v0.1.0 // indirect
)

replace (
	github.com/philippgille/gokv => ../
	github.com/philippgille/gokv/encoding => ../encoding
	github.com/philippgille/gokv/sql => ../sql
	github.com/philippgille/gokv/test => ../test
	github.com/philippgille/gokv/util => ../util
)
