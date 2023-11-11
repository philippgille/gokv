$workingDir = $(pwd).Path

# We don't want to update transitive dependencies, so instead of using `go get -u` we use
# `go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)`
# as suggested in https://github.com/golang/go/issues/28424#issuecomment-1101896499.

# Helper packages
$array = @("encoding","sql","test", "util")
foreach ($moduleName in $array){
    echo "updating $moduleName"
    cd "$PSScriptRoot/../$moduleName"; go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all); go mod tidy; cd $workingDir
}

# Implementations
cat "$PSScriptRoot/implementations" | foreach {
    echo "updating $_"
    cd "$PSScriptRoot/../$_"; go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all); go mod tidy; cd $workingDir
}

# Examples
echo "updating examples"
cd "$PSScriptRoot/../examples/redis"; go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all); go mod tidy; cd $workingDir
cd "$PSScriptRoot/../examples/proto_encoding"; go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all); go mod tidy; cd $workingDir
