#!/bin/bash

IP=$(ip addr show eth0 | grep "inet\b" | awk '{print $2}' | cut -d/ -f1)
echo $IP
PORT=10250
KUBECONFIG=~/.kube/config krustlet-wasi --node-ip=$IP --port=10250