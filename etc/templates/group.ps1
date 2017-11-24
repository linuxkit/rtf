# SUMMARY: 
# LABELS:
# For the top level group.ps1 also specify a 'NAME:' comment

# Print each line executed
Set-PSDebug -Trace 2

# Source libraries. Uncomment if needed/defined
#. $env:RT_LIB

# returns from functions in posh seem to contain all the output, so we
# use a global variable to get the return value.
$res = 0

function GroupInit([REF]$res) {
    # Group initialisation code goes here
    $res.Value = 0
}

function GroupDeinit([REF]$res) {
    # Group de-initialisation code goes here
    $res.Value = 0
}

$CMD=$args[0]
Switch ($CMD) {
    'init'    { GroupInit $res }
    'deinit'  { GroupDeinit $res }
}

exit $res
