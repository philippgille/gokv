module github.com/philippgille/gokv/consul

go 1.12

require (
	github.com/armon/go-metrics v0.0.0-20190430140413-ec5e00d3c878 // indirect
	github.com/go-test/deep v1.0.4 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/hashicorp/consul/api v1.2.0
	github.com/hashicorp/go-immutable-radix v1.1.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/go-rootcerts v1.0.1 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/hashicorp/memberlist v0.1.5 // indirect
	github.com/hashicorp/serf v0.8.4 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/miekg/dns v1.1.19 // indirect
	github.com/philippgille/gokv v0.5.0
	github.com/stretchr/testify v1.4.0 // indirect
	golang.org/x/crypto v0.0.0-20190927123631-a832865fa7ad // indirect
	golang.org/x/net v0.0.0-20190926025831-c00fd9afed17 // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/sys v0.0.0-20190927073244-c990c680b611 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
