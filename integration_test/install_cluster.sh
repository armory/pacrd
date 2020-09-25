#!/bin/bash

####### This code is to make it work on drydock
# Create a cluster
# For some reason region needs o exists
# echo "Setting up WS_DEFAULT_REGION=us-west-2"
# export AWS_DEFAULT_REGION=us-west-2

# echo "Executing drydock -b"
# drydock init -b pacrd_integration_test

# echo "Executing ensure-cluster"
# drydock ensure-cluster

# echo "Executing drydock apply-connect"
# drydock apply-connect -- --port-forward service/spin-gate 8084 --port-forward service/spin-deck 9000
 

echo "Starting docker daemon"
$(nohup dockerd-entrypoint.sh) &

echo "Sleep 20 seconds"
sleep 20

echo "Creating cluster"
k3d cluster create mycluster

echo "Installing kubeconfig in /home/spinnaker/.kube/config because clouddriver wants it like that"
mkdir -p /home/spinnaker/.kube/
cp ~/.kube/config /home/spinnaker/.kube/config

echo "Downloading repo"
git clone https://github.com/armory/spinnaker-kustomize-patches.git

echo "Entering in repo"
cd spinnaker-kustomize-patches

echo "Executing deployment of operator"
$(nohup ./deploy.sh) &

echo "Waiting 5 minutes"
sleep 300

echo "Changing to namespace spinnaker"
kubectl config set-context --current --namespace=spinnaker

echo "Waiting for all the spinnaker services"
declare -i FOUND=0
while true; do
    FOUND=0
    while read p; do
        read READY_STATUS <<< $p
        if [ -z $READY_STATUS ] || [ $READY_STATUS != "1/1" ]; then
            echo "Found a pod in state:" $READY_STATUS
            FOUND=1
            break
        fi
    done <<< $(kubectl get po |  awk '{ print $2 }' | tail -n +2)
    if [[ $FOUND == 0 ]]; then
        break
    fi
echo "Will query in 30 seconds"
sleep 30
done


echo "Port forward gate and deck"
$(nohup kubectl port-forward services/spin-gate 8084:8084) &
$(nohup kubectl port-forward services/spin-deck 9000:9000) &


echo "Leaving directory"
cd ..