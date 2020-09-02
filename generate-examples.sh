#!/bin/bash -e

if [ ! -d "env" ]; then
    python3 -m venv env
fi

function deactivateEnv {
    rv=$?
    deactivate
    exit $?
}

trap "deactivateEnv" EXIT

source env/bin/activate

# uninstall cloudformation-cli-go-plugin if it exists
pip3 show cloudformation-cli-go-plugin
SHOW_RV=$?

if [ "$SHOW_RV" == "0" ]; then
    pip3 uninstall -y cloudformation-cli-go-plugin
fi

pip3 install -e .

EXAMPLES=( github-repo )
for EXAMPLE in "${EXAMPLES[@]}"
do
    cd examples/$EXAMPLE
    cfn generate
done
