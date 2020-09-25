#!/bin/bash

echo "Creating install_pacrd_patched directory"
mkdir install_pacrd_patched

# echo "Download CRD for PaCRD-$PACRD_VERSION --> https://engineering.armory.io/manifests/pacrd-$PACRD_VERSION.yaml"
# curl -fsSL https://engineering.armory.io/manifests/pacrd-$PACRD_VERSION.yaml > install_pacrd_patched/pacrd.yaml
echo "Using pacrd.yaml generated with make integration-test"
cp pacrd.yaml install_pacrd_patched/pacrd.yaml

echo "Changing namespace in kustomization_template.yaml"
sed "s/<namespace>/$SPIN_NAMESPACE/g" install_pacrd/kustomization_template.yaml > install_pacrd_patched/kustomization.yaml

echo "Changing namespace in patch_template.yaml"
sed "s/<namespace>/$SPIN_NAMESPACE/g" install_pacrd/patch_template.yaml > install_pacrd_patched/patched.yaml

echo "Kustomize on the changed template files for PaCRD and saving it in $KUSTOMIZED_PATH"
kustomize build  install_pacrd_patched/ > $KUSTOMIZED_PATH

# Spin-cli sends the yaml declarations as json so this code is to convert the pacrd install yaml and tests to json
# PaCRD install file
# Convert to json since deploy stage changes the yaml to json
echo "Converting kustomized yaml to base64 to avoid issues with spacing while sending pipelines"
cat $KUSTOMIZED_PATH | base64 | tr -d '\n' > $KUSTOMIZED_PATCHED_BASE64

# This code merges all the files in pacrd-test-samples with --- separator and then converts them to json
# Same for test cases in PaCRD folder
# ../config/samples/pacrd-test-samples/
echo "Reading all files form pacrd-test-samples/ and concatenating them with --- since those should be yamls"
for file in pacrd-test-samples/*; do
cat $file >> $PACRD_TEST_SAMPLES
printf '\n\n---\n\n' >> $PACRD_TEST_SAMPLES
done

echo "Converting the concatenated file to base64 to avoid issues with spacing while sending pipelines"
cat $PACRD_TEST_SAMPLES | base64 | tr -d '\n' > $PACRD_TEST_SAMPLES_BASE64

# Same for integration test in PaCRD folder
# pacrd-integration-test-samples/*
echo "Reading all files form pacrd-integration-test-samples/ and concatenating them with --- since those should be yamls"
for file in pacrd-integration-test-samples/*; do
cat $file >> $PACRD_INTEGRATION_TEST_SAMPLES
printf '\n\n---\n\n' >> $PACRD_INTEGRATION_TEST_SAMPLES
done

echo "Converting the concatenated file to base64 to avoid issues with spacing while sending pipelines"
cat $PACRD_INTEGRATION_TEST_SAMPLES | base64 | tr -d '\n' > $PACRD_INTEGRATION_TEST_SAMPLES_BASE64

# This needs to exists because then we make some jq to execute them
echo "Creating an additional json convertion from previous yaml file for pacrd-integration-test-samples, this is used in future to know which pipelines to execute"
echo -e 'import os, sys, yaml, json\narr =[]\nfor c in yaml.safe_load_all(open(os.getenv("PACRD_INTEGRATION_TEST_SAMPLES"), "r").read()): arr.append(c) if c != None else None\nprint(json.dumps(arr))' | python3 > $PACRD_INTEGRATION_TEST_SAMPLES_JSON