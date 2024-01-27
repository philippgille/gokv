module github.com/philippgille/gokv/mysql

go 1.20

require (
	github.com/go-sql-driver/mysql v1.7.1
	github.com/philippgille/gokv/encoding v0.6.0
	github.com/philippgille/gokv/sql v0.6.0
	github.com/philippgille/gokv/test v0.6.0
)

require (
	github.com/go-test/deep v1.1.0 // indirect
	github.com/philippgille/gokv v0.6.0 // indirect
	github.com/philippgille/gokv/util v0.6.0 // indirect
)

replace (
	github.com/philippgille/gokv => ../
	github.com/philippgille/gokv/encoding => ../encoding
	github.com/philippgille/gokv/sql => ../sql
	github.com/philippgille/gokv/test => ../test
	github.com/philippgille/gokv/util => ../util
)
