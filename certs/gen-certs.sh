#!/usr/bin/env bash

# Generate the CA cert and private key for our self signed cert
openssl req -nodes -new -x509 -keyout ca.key -days 3650 -out ca.crt -subj "/CN=ingress-validator.default.svc"
## Generate the private key for the webhook server

# openssl genrsa -out ingress-validator-tls.key 2048

openssl req -newkey rsa:2048 -nodes -keyout ingress-validator-tls.key -subj "/CN=ingress-validator.default.svc" -out ingress-validator-tls.csr

openssl x509 -req -extfile <(printf "subjectAltName=DNS:ingress-validator.default.svc,DNS:ingress-validator.default,DNS:ingress-validator") -days 3650 -in ingress-validator-tls.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out ingress-validator-tls.crt
