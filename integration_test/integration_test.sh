#!/bin/bash

export KUSTOMIZED_PATH="install_pacrd_patched/kustomized.yaml"
export KUSTOMIZED_PATCHED_BASE64="install_pacrd_patched/kustomized.base64"
export PACRD_TEST_SAMPLES="install_pacrd_patched/combined_tests.yaml"
export PACRD_TEST_SAMPLES_BASE64="install_pacrd_patched/combined_tests.base64"
export PACRD_INTEGRATION_TEST_SAMPLES="install_pacrd_patched/combined_integration_tests.yaml"
export PACRD_INTEGRATION_TEST_SAMPLES_JSON="install_pacrd_patched/combined_integration_tests.json"
export PACRD_INTEGRATION_TEST_SAMPLES_BASE64="install_pacrd_patched/combined_integration_tests.base64"
export SPIN_NAMESPACE=spinnaker

if [ -z "$PACRD_VERSION" ]; then
      export PACRD_VERSION=0.10.1
fi

unset DOCKER_HOST

echo "Running install_cluster.sh"
bash ./install_cluster.sh

echo "Running kubernetes_account.sh"
bash ./kubernetes_account.sh

echo "Running configure_spin_cli.sh"
bash ./configure_spin_cli.sh

echo "Running kustomize_pacrd.sh"
bash ./kustomize_pacrd.sh

echo "Running patch_spin_cli_files.sh"
bash ./patch_spin_cli_files.sh

echo "Running spin_cli_execute_pipeline.sh"
bash ./spin_cli_execute_pipeline.sh
