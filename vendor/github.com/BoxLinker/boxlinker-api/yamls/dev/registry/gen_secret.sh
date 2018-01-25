#!/bin/bash

kubectl delete secret registry-server-config --namespace=boxlinker || true
kubectl create secret generic registry-server-config --from-file ./auth_config.yml --namespace=boxlinker