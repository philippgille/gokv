module github.com/philippgille/gokv/tablestorage

go 1.12

require (
	github.com/Azure/azure-sdk-for-go v33.4.0+incompatible
	github.com/Azure/go-autorest/autorest v0.9.1 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.6.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/dnaeon/go-vcr v1.0.1 // indirect
	github.com/go-test/deep v1.0.4 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/philippgille/gokv v0.5.0
	github.com/satori/go.uuid v1.2.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
