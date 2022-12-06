#!/bin/bash
# 환경설정용 환경변수 
# export KUBECONFIG=/root/kubeconfig/ck1-kubeconfig.yaml
export KUBECONFIG=/root/kubeconfig/vsphere-kubeconfig.yaml

## uninstall crd + operator undeploy
make undeploy-test
