# CLI for GoKV

## Usage

```
$ go run main.go set foo bar
Using config file: /Users/rijojohn/GolandProjects/gokv/cli/example/gokv.yaml
Value added to store

$ go run main.go get foo
Using config file: /Users/rijojohn/GolandProjects/gokv/cli/example/gokv.yaml
Retrieved Value: bar


$ go run main.go delete foo
Using config file: /Users/rijojohn/GolandProjects/gokv/cli/example/gokv.yaml
Key foo has been deleted.%
```

Currently using config file ./example/gokv.yaml
