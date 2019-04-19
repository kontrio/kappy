#!/usr/bin/env bash

set -e

machine=$(echo "$(uname -s)" | tr '[:upper:]' '[:lower:]')
arch=$(echo "$(uname -m)" | sed s/^x86_64/amd64/g)

TARGET=${machine}_${arch}
KAPPY_BIN_DIR=./dist/$TARGET/

kubectl proxy 2>&1 > /dev/stderr &

kubectl get pods --all-namespaces


kappy_deploy() {
  set -e
  $KAPPY_BIN_DIR/kappy version
  $KAPPY_BIN_DIR/kappy deploy $1 --file ./test/kappy.test.$1.yaml --version 1 --verbose
}

test_internalservice() {
  kappy_deploy internalservice

  result=$(curl -s http://127.0.0.1:8001/api/v1/namespaces/internalservice/services/http:echo-test:80/proxy/)

  if [ "$result" != "\"received request\"" ];then 
    echo "Request to internally deployed service failed"
    return 1
  fi

  kubectl delete namespace/internalservice 
  echo "PASSED"
}

test_externalservice() {
  kappy_deploy externalservice

  set -x

  kubectl get pods --namespace externalservice
  kubectl get services --namespace externalservice
  kubectl get endpoints echo-test --namespace externalservice
  kubectl get ingresses --namespace externalservice
  kubectl get certificates --namespace externalservice
  sleep 5

  resultinternal=$(curl -s http://127.0.0.1:8001/api/v1/namespaces/externalservice/services/http:echo-test:80/proxy/)
  if [ "$resultinternal" != "\"received request\"" ];then 
    echo "Request to internally deployed service failed"
    return 1
  fi

  result=$(curl https://$(minikube ip)/ -H "Host: testecho.kappycitests.kontr.io" -k)
  if [ "$result" != "\"received request\"" ];then 
    echo "Request to externally deployed service failed"
    return 1
  fi
  
  kubectl delete namespace/externalservice 
  echo "PASSED"
}

test_internalservice
test_externalservice

