#!/usr/bin/env bash
# Copyright (c) 2022, Linaro. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

# Exit on command failure
set -o errexit

# Fail on unset variable
# Use "${VARNAME-}" instead of "$VARNAME" to access unset variable(s)
set -o nounset

# Enable debug mode if $TRACE is set
# To enable, run with: "env TRACE=1 ./setup-bootstrap.sh"
if [[ "${TRACE-0}" == "1" ]]; then
    set -o xtrace
fi

# Check if the first arg is -h or --help
if [[ "${1-}" =~ ^-*h(elp)?$ ]]; then
    echo "Usage: ./setup-bootstrap.sh

Generates a device class certicate and keypair that enables client devices to
connect to the bootstrap server during device provisioning, or to interact with
the REST API using mutual TLS authentication.

The device class certificate is signed by the bootstrap server's CA key, and
has a specific string in the subject line that is verified by the bootstrap
server during the TLS handshake.

This certificate and private key must be available on client devices. If the
certificate and matching private key is not available on the client device, the
server will refuse the TLS connection.

The following files are placed them in the 'certs' folder:

- certs/BOOTSTRAP.crt      Device class certificate for the bootstrap server
- certs/BOOTSTRAP.key      Private key associated with BOOTSTRAP.crt
- certs/bootstrap_crt.txt  A C string copy of BOOTSTRAP.crt (as a convenience)
- certs/bootstrap_key.txt  A C string copy of BOOTSTRAP.key (as a convenience)

You can view the content of the certificate via:

   $ openssl x509 -in certs/BOOTSTRAP.crt -noout -text
"
    exit
fi

# Generate a keypair for initial device provisioning.  The idea is
# that this private key and certificate can be programmed into the
# devices, within a class, at factory time.  This will allow these
# devices to request certificates later to authenticate themselves.

# For now, write them to a single name
if [ -f certs/BOOTSTRAP.crt ] || [ -f certs/BOOTSTRAP.key ];
then
	echo "Device Class certificates seem to already be present."
	exit 1
fi

if [ ! -f certs/CA.key ] || [ ! -f certs/CA.crt ];
then
	echo "CA cert needs to be created first."
	echo ""
	echo "Run 'setup-ca.sh' before using this script."
	exit 1
fi

mkdir -p certs

# This is a simple keypair, where the certificate will be known by the
# server.
openssl ecparam -name prime256v1 -genkey -out certs/BOOTSTRAP.key

# Generate the CSR
openssl req -new -sha256 -key certs/BOOTSTRAP.key \
	-out certs/BOOTSTRAP.csr \
	-subj '/O=Linaro, LTD/CN=bootstrap-register-1/OU=LinaroCA Bootstrap Cert'

# Sign it with our CA cert
openssl x509 -req -sha256 \
	-CA certs/CA.crt \
	-CAkey certs/CA.key \
	-days 3560 \
	-CAcreateserial \
	-CAserial certs/CA.srl \
	-in certs/BOOTSTRAP.csr \
	-out certs/BOOTSTRAP.crt

rm certs/BOOTSTRAP.csr

# Convert the public and private keys into C code so they can be
# included into the app.
sed 's/.*/"&\\r\\n"/' certs/BOOTSTRAP.crt > certs/bootstrap_crt.txt
openssl ec -in certs/BOOTSTRAP.key -outform DER |
	xxd -i > certs/bootstrap_key.txt
