#!/bin/env sh

docker build -t reg.kolla.org/library/ingress-validator:v1.0.0 .

docker save -o 111 reg.kolla.org/library/ingress-validator:v1.0.0

ctr -n k8s.io image remove reg.kolla.org/library/ingress-validator:v1.0.0

ctr -n k8s.io image import 111

rm -f 111