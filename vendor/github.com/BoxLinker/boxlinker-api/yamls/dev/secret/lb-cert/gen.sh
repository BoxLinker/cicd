#!/bin/bash

kubectl create secret tls lb-cert --cert=./ca.crt --key=./ca.key
