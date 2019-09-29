module github.com/philippgille/gokv/sql

go 1.12

require (
	github.com/philippgille/gokv/encoding v0.0.0-20190929161440-07d380f5709c
	github.com/philippgille/gokv/util v0.0.0-20190929161440-07d380f5709c
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
