#!/bin/bash

HOSTNAME=localhost

# Build linaroca
go build

# Clean up previous artifacts
rm CA.* SERVER.* USER.*

# Clean up the database (this is for testing), since once we've
# removed the CA.crt, those certs are fairly meanmingless.
rm CADB.db

# Generate private key for the HTTP server
openssl ecparam -name secp256r1 -genkey -out SERVER.key

# Generate a self-signed certificate
openssl req -new -x509 -sha256 -days 3650 -key SERVER.key -out SERVER.crt \
        -subj "/O=Linaro, LTD/CN=$HOSTNAME"

# Generate the CA certificate
./linaroca cakey generate

# Generate private user key
openssl ecparam -name prime256v1 -genkey -out USER.key

# Generate a user CSR with a random UUID
openssl req -new -key USER.key -out USER.csr \
    -subj "/O=$HOSTNAME/CN=$(uuidgen | tr '[:upper:]' '[:lower:]')"

# Convert CSR to JSON
go run make_csr_json.go

echo "HTTP cert, CA cert and USER CSR generated."

# Start linaroca
./linaroca server start -p 1443

# At this point we can submit the CSR via:
# $ wget --ca-certificate=SERVER.crt \
#     --post-file USER.json \
#     https://localhost:1443/api/v1/cr \
#     -O USER.cr
#
# Convert it to DER then PEM:
# $ jq -r '.Cert' < USER.cr | base64 --decode > USER.der
# $ openssl x509 -in USER.der -inform DER -out USER.crt -outform PEM
#
# Then, verify it:
# $ openssl x509 -in USER.crt -noout -text
