module github.com/philippgille/gokv/mysql

go 1.12

require (
	github.com/go-sql-driver/mysql v1.4.1
	github.com/go-test/deep v1.0.4 // indirect
	github.com/philippgille/gokv v0.5.0
	google.golang.org/appengine v1.6.4 // indirect
)

// Required to fix ambiguous import path
replace github.com/philippgille/gokv => ../
