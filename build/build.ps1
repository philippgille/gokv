$workingDir = $(pwd).Path

# Interface module
cd "$PSScriptRoot/.."; go build -v; cd $workingDir

# Helper packages
$array = @("encoding","sql","test", "util")
foreach ($moduleName in $array){
    echo "building $moduleName"
    cd "$PSScriptRoot/../$moduleName"; go build -v; cd $workingDir
}

# Implementations
cat "$PSScriptRoot/implementations" | foreach {
    echo "building $_"
    cd "$PSScriptRoot/../$_"; go build -v; cd $workingDir
}

# Examples
echo "building examples"
cd "$PSScriptRoot/../examples"; go build -v; cd $workingDir
