#! /bin/bash

# Hostname for certificates.  'localhost' is good for testing,
# especially if you are behind a NAT.
HOSTNAME=localhost

# Setup the Certificate Authority and server certificates.  In
# general, this should be run once, to create these initial
# certificates, for development.

if [ -f CA.crt -o -f CA.key -o -f CADB.db \
	-o -f SERVER.crt -o -f SERVER.key ];
then
	echo "Server/CA certificates seem to already be present."
	exit 1
fi

# Build the application.
go build || exit 1

# Generate a self-signed certificate for the REST API (HTTP).
openssl ecparam -name secp256r1 -genkey -out SERVER.key
openssl req -new -x509 -sha256 -days 3650 -key SERVER.key -out SERVER.crt \
        -subj "/O=Linaro, LTD/CN=$HOSTNAME"

# Generate the CA certificate
./linaroca cakey generate
