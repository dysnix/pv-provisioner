#!/usr/bin/env bash
kubectl create -f pvc-from-snapshot.yaml
kubectl create -f pod-from-snapshot.yaml