#!/usr/bin/env bash
docker build -t asia.gcr.io/test-chainstack/pv-provisioner ./
docker push asia.gcr.io/test-chainstack/pv-provisioner