module github.com/philippgille/gokv/bigcache

go 1.12

require (
	github.com/allegro/bigcache v1.2.1
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-test/deep v1.0.4 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/philippgille/gokv v0.5.0
	github.com/stretchr/testify v1.4.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
