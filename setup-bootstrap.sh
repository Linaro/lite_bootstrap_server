#! /bin/bash

set -e

# Generate a keypair for initial device provisioning.  The idea is
# that this private key and certificate can be programmed into the
# devices, within a class, at factory time.  This will allow these
# devices to request certificates later to authenticate themselves.

# For now, write them to a single name
if [ -f certs/BOOTSTRAP.crt -o -f certs/BOOTSTRAP.key ];
then
	echo "Device Class certificates seem to already be present."
	exit 1
fi

if [ ! -f certs/CA.key -o ! -f certs/CA.crt ];
then
	echo "CA cert needs to be created first."
	exit 1
fi

mkdir -p certs

# This is a simple keypair, where the certificate will be known by the
# server.
openssl ecparam -name prime256v1 -genkey -out certs/BOOTSTRAP.key

# Generate the CSR
openssl req -new -sha256 -key certs/BOOTSTRAP.key \
	-out certs/BOOTSTRAP.csr \
	-subj '/O=Linaro, LTD/CN=bootstrap-register-1'

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
