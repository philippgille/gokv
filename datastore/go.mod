module github.com/philippgille/gokv/datastore

go 1.12

require (
	cloud.google.com/go/datastore v1.0.0
	cloud.google.com/go/storage v1.0.0 // indirect
	github.com/go-test/deep v1.0.4 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/philippgille/gokv v0.5.0
	go.opencensus.io v0.22.1 // indirect
	golang.org/x/exp v0.0.0-20190927203820-447a159532ef // indirect
	golang.org/x/net v0.0.0-20190926025831-c00fd9afed17 // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/sys v0.0.0-20190927073244-c990c680b611 // indirect
	google.golang.org/api v0.10.0
	google.golang.org/appengine v1.6.4 // indirect
	google.golang.org/genproto v0.0.0-20190927181202-20e1ac93f88c // indirect
	google.golang.org/grpc v1.24.0 // indirect
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
