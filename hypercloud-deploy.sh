#!/bin/bash
# 환경설정용 환경변수
export KUBECONFIG=/root/kubeconfig/ck1-kubeconfig.yaml # 실환경 
# export KUBECONFIG=/root/kubeconfig/vsphere-kubeconfig.yaml 

## crd 설치 및 operator deploy
export IMG=192.168.9.12:5000/hypercloud-multi-operator-dev:$1

# image build and push
./push.sh $1

# kubectl create ns hypercloud5-system
make deploy-test
