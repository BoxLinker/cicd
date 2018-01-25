#!/bin/bash

kubectl delete secret env-registry-watcher --namespace=boxlinker
kubectl create secret generic env-registry-watcher --from-file=`pwd`/env.yml --namespace=boxlinker