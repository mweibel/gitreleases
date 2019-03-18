#!/bin/bash
#
# green/blue deployment
# From: https://www.ianlewis.org/en/bluegreen-deployments-kubernetes
#

DEPLOYMENT=$1
PREV_DEPLOYMENT=`kubectl get deployments -n gitreleases -o=jsonpath='{.items[0].metadata.name}'`
cat k8s/deployment.yml | sed 's/{{TAG}}/'${DEPLOYMENT}'/g'| kubectl apply -f -

# Wait until the Deployment is ready by checking the MinimumReplicasAvailable condition.
READY=$(kubectl get deploy -n gitreleases gitreleases-$DEPLOYMENT -o json | jq '.status.conditions[] | select(.reason == "MinimumReplicasAvailable") | .status' | tr -d '"')

while [[ "$READY" != "True" ]]; do
    READY=$(kubectl get deploy -n gitreleases gitreleases-$DEPLOYMENT -o json | jq '.status.conditions[] | select(.reason == "MinimumReplicasAvailable") | .status' | tr -d '"')
    sleep 5
done

cat k8s/service.yml | sed 's/{{TAG}}/'${DEPLOYMENT}'/g'| kubectl apply -f -

kubectl delete deployment -n gitreleases $PREV_DEPLOYMENT
