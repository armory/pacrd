#!/bin/bash

# Patch pipeline json
echo "Add test examples manifests to pipeline template that is going to be send to gate"
jq ".stages[0].manifestArtifact.reference = \"$(cat $PACRD_TEST_SAMPLES_BASE64)\"" spin_cli_pipe/pipeline.json > spin_cli_pipe/pipeline_patched.json

echo "Add PaCRD install manifests to pipeline template that is going to be send to gate"
cat <<< $(jq ".stages[2].manifestArtifact.reference = \"$(cat $KUSTOMIZED_PATCHED_BASE64)\"" spin_cli_pipe/pipeline_patched.json) > spin_cli_pipe/pipeline_patched.json

echo "Add test integration examples manifests to pipeline template that is going to be send to gate"
cat <<< $(jq ".stages[3].manifestArtifact.reference = \"$(cat $PACRD_INTEGRATION_TEST_SAMPLES_BASE64)\"" spin_cli_pipe/pipeline_patched.json) > spin_cli_pipe/pipeline_patched.json

echo "Add application and name of the pipeline that is going to be send to gate"
cat <<< $(jq '.name="install-pacrd-and-tests"' spin_cli_pipe/pipeline_patched.json) > spin_cli_pipe/pipeline_patched.json

cat <<< $(jq '.application="testpacrd"' spin_cli_pipe/pipeline_patched.json) > spin_cli_pipe/pipeline_patched.json

echo "Create application testpacrd"
spin application save --file spin_cli_pipe/application.json

echo "Create pipeline install-pacrd-and-tests to install PaCRD"
spin pipeline save -f spin_cli_pipe/pipeline_patched.json