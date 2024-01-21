Protobuf encoding example
=========================

For this example we already generated the Go file from the `.proto` file, so you can run it with `go run .`.

Generate Go files
-----------------

If you want to do the generation yourself, you have to:

1. Install protoc: <https://github.com/protocolbuffers/protobuf#protobuf-compiler-installation>
   - We tested with [v25.0](https://github.com/protocolbuffers/protobuf/releases/tag/v25.0)
2. Install the Go protocol buffers plugin: `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`
   - We tested with [v1.31.0](https://pkg.go.dev/google.golang.org/protobuf@v1.31.0/cmd/protoc-gen-go)
3. Generate Go files from `.proto` files: `protoc --go_out=./tutorialpb --go_opt=paths=source_relative ./addressbook.proto`

For a complete official protobuf tutorial see <https://protobuf.dev/getting-started/gotutorial/>.

Inspect the raw bytes
---------------------

If you want to see how the stored value looks like, in its raw, encoded form, you can add the following at the end of the `main` function:

```go
// Just to demonstrate the raw encoded value, we can print the raw bytes as string
// using reflection to access the unexported field.
reflectedField := reflect.ValueOf(&store).Elem().FieldByName("m")
reflectedField = reflect.NewAt(reflectedField.Type(), unsafe.Pointer(reflectedField.UnsafeAddr())).Elem()
m := reflectedField.Interface().(map[string][]byte)
rawVal := m["foo123"]
fmt.Printf("Raw value: %s\n", rawVal)
// Prints:
// Raw value:
// John Doe����johndoe@example.com"
//
// 0123-456789"
//
// 0987-654321*
//            󟛪����
```

The reflection usage is only required because the example uses the `gomap` in-memory store. If it was using Redis or another remote store, you could just create a plain (non-`gokv`) client and retrieve the value with that. It would lead to the same result.
