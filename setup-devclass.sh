#! /bin/bash

set -e

# Generate a keypair for initial device provisioning.  The idea is
# that this private key and certificate can be programmed into the
# devices, within a class, at factory time.  This will allow these
# devices to request certificates later to authenticate themselves.

# For now, write them to a single name
if [ -f certs/DEVCLASS.crt -o -f certs/DEVCLASS.key ];
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
openssl ecparam -name prime256v1 -genkey -out certs/DEVCLASS.key

# Generate the CSR
openssl req -new -sha256 -key certs/DEVCLASS.key \
	-out certs/DEVCLASS.csr \
	-subj '/O=Linaro, LTD/CN=devclass-register-1'

# Sign it with our CA cert
openssl x509 -req -sha256 \
	-CA certs/CA.crt \
	-CAkey certs/CA.key \
	-days 3560 \
	-CAcreateserial \
	-CAserial certs/CA.srl \
	-in certs/DEVCLASS.csr \
	-out certs/DEVCLASS.crt

rm certs/DEVCLASS.csr

# Convert the public and private keys into C code so they can be
# included into the app.
sed 's/.*/"&\\r\\n"/' certs/DEVCLASS.crt > certs/devclass_crt.txt
openssl ec -in certs/DEVCLASS.key -outform DER |
	xxd -i > certs/devclass_key.txt
