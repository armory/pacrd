#!/bin/bash

echo "Getting pipeline id for install-pacrd-and-tests in testpacrd"
PACRD_PIPELINE_ID=$(spin pipeline get --application testpacrd --name install-pacrd-and-tests | jq --raw-output '.id')

echo "Executing pipeline install-pacrd-and-tests"
spin pipeline execute --application testpacrd --name install-pacrd-and-tests

echo "Sleep 5 seconds waiting for execution"
sleep 5

echo "Querying execution status until finishes running"
while true; do
  if [[ $(spin pipeline execution list --pipeline-id $PACRD_PIPELINE_ID | jq --raw-output '.[0].status') != "RUNNING" ]]; then
    break
  fi
  echo "Pipeline install-pacrd-and-tests is still running..."
  sleep 10
done

# If finished status is not SUCCEEDED then fail
if [[ $(spin pipeline execution list --pipeline-id $PACRD_PIPELINE_ID | jq --raw-output '.[0].status') != "SUCCEEDED" ]]; then
  echo "Pipeline install-pacrd-and-tests failed!"
  spin pipeline execution list --pipeline-id $PACRD_PIPELINE_ID >&2
  exit 1
fi

echo "Creating file app-pipeline-relationship.txt that will save app-pipeline relationships from combined_integration_tests yamls as application pipeline so we can execute them next"
jq --raw-output '.[] |select( .kind == "Pipeline" ) | .spec.application + " " + .metadata.name' install_pacrd_patched/combined_integration_tests.json > install_pacrd_patched/app-pipeline-relationship.txt

while read p; do
  read CURRENT_APPLICATION CURRENT_PIPELINE <<< $p
  echo "\n\nStarting to execute pipeline with this values:"
  echo "CURRENT_APPLICATION=$CURRENT_APPLICATION"
  echo "CURRENT_PIPELINE=$CURRENT_PIPELINE"

  # Get pipeline id
  PACRD_PIPELINE_ID=$(spin pipeline get --application $CURRENT_APPLICATION --name $CURRENT_PIPELINE | jq --raw-output '.id')
  
  # Execute pipeline
  spin pipeline execute --application $CURRENT_APPLICATION --name $CURRENT_PIPELINE
  
  # Wait 5 seconds
  sleep 5

  # Query execution status until finishes running
  while true; do
    if [[ $(spin pipeline execution list --pipeline-id $PACRD_PIPELINE_ID | jq --raw-output '.[0].status') != "RUNNING" ]]; then
        break
    fi
    echo "Pipeline $CURRENT_PIPELINE is still running..."
    sleep 10
  done

  # If finished status is not SUCCEEDED then fail
  if [[ $(spin pipeline execution list --pipeline-id $PACRD_PIPELINE_ID | jq --raw-output '.[0].status') != "SUCCEEDED" ]]; then
    echo  "Execution failed for CURRENT_APPLICATION=$CURRENT_APPLICATION CURRENT_PIPELINE=$CURRENT_PIPELINE"
    spin pipeline execution list --pipeline-id $PACRD_PIPELINE_ID >&2
    exit 1
  fi

# Do this for all the pipelines previously parsed from combined_integration_tests
done < install_pacrd_patched/app-pipeline-relationship.txt

echo "Integration tests finished successfully!"