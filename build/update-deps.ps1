$workingDir = $(pwd).Path

# Note: To update only the direct dependencies, use:
# go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)

# Helper packages
$array = @("encoding","sql","test", "util")
foreach ($moduleName in $array){
    echo "updating $moduleName"
    cd "$PSScriptRoot/../$moduleName"; go get -u -t; go mod tidy; cd $workingDir
}

# Implementations
cat "$PSScriptRoot/implementations" | foreach {
    echo "updating $_"
    cd "$PSScriptRoot/../$_"; go get -u -t; go mod tidy; cd $workingDir
}

# Examples
echo "updating examples"
cd "$PSScriptRoot/../examples"; go get -u -t; go mod tidy; cd $workingDir
