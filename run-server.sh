#!/bin/bash

HOSTNAME=localhost

# Build linaroca
go build || exit 1

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

# Start linaroca
./linaroca server start -p 1443
