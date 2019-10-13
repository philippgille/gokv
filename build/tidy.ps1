$workingDir = $(pwd).Path

# Interface module
cd "$PSScriptRoot/.."; go mod tidy; cd $workingDir

# Helper packages
$array = @("encoding","sql","test", "util")
foreach ($moduleName in $array){
    echo "tidying $moduleName"
    cd "$PSScriptRoot/../$moduleName"; go mod tidy; cd $workingDir
}

# Implementations
cat "$PSScriptRoot/implementations" | foreach {
    echo "tidying $_"
    cd "$PSScriptRoot/../$_"; go mod tidy; cd $workingDir
}

# Examples
echo "tidying examples"
cd "$PSScriptRoot/../examples"; go mod tidy; cd $workingDir
