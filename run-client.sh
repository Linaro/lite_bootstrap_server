#!/bin/bash

HOSTNAME=localhost

# Generate private user key
openssl ecparam -name prime256v1 -genkey -out USER.key

# Generate a user CSR with a random UUID
openssl req -new -key USER.key -out USER.csr \
    -subj "/O=$HOSTNAME/CN=$(uuidgen | tr '[:upper:]' '[:lower:]')/OU=LinaroCA Device Cert - Signing"

# Convert CSR to JSON
go run make_csr_json.go

echo "HTTP cert, CA cert and USER CSR generated."

# At this point we can submit the CSR via:
wget --ca-certificate=SERVER.crt \
  --post-file USER.json \
  https://localhost:1443/api/v1/cr \
  -O USER.rsp

# Convert it to DER then PEM:
jq -r '.Cert' < USER.rsp | base64 --decode > USER.der
openssl x509 -in USER.der -inform DER -out USER.crt -outform PEM

# Then, display it:
openssl x509 -in USER.crt -noout -text
