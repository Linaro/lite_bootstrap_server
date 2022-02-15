#! /bin/bash

# Exit on any error
set -e

# Create a new testing device, and register it.

HOSTNAME=localhost

# Generate a device ID.  BSD's uuidgen outputs uppercase, so conver
# that here.
DEVID=$(uuidgen | tr '[:upper:]' '[:lower:]')

echo New device: $DEVID

# Generate a private user key for this device.
openssl ecparam -name prime256v1 -genkey -out $DEVID.key

# Generate the CSR for this key.
openssl req -new \
	-key $DEVID.key \
	-out $DEVID.csr \
	-subj "/O=$HOSTNAME/CN=$DEVID/OU=LinaroCA Device Cert - Signing"

# Convert this CSR to json.
go run make_csr_json.go -in $DEVID.csr -out $DEVID.json

# Submit the CSR.
wget --ca-certificate=SERVER.crt \
	--post-file $DEVID.json \
	https://localhost:1443/api/v1/cr \
	-O $DEVID.rsp

# Convert it to DER then PEM.
jq -r ".Cert" < $DEVID.rsp | base64 --decode > $DEVID.der
openssl x509 -in $DEVID.der -inform DER -out $DEVID.crt -outform PEM

# Display the certificate
openssl x509 -in $DEVID.crt -noout -text
