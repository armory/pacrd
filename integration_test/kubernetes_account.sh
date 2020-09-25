#!/bin/bash

echo "Change namespace value in rbac template file"
sed "s/<namespace>/$SPIN_NAMESPACE/g" kubernetes_account/rbac.yaml > kubernetes_account_patch/rbac.yaml

echo "kubectl apply on rbac file"
kubectl apply -f kubernetes_account_patch/

echo "kubectl patch on SpinnakerService spinnaker to add namespace since its needed to deploy"
kubectl patch SpinnakerService spinnaker --type merge  --patch "$(cat kubernetes_spinsvc_patch/patch.json)"

echo "Waiting 1 minute"
sleep 60

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
