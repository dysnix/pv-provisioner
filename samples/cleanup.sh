#!/usr/bin/env bash
kubectl delete -f pod-from-snapshot.yaml
kubectl delete -f pod-standart.yaml
kubectl delete -f pvc-from-snapshot.yaml
kubectl delete -f pvc-standart.yaml
kubectl delete pv --all