#!/usr/bin/env bash
# Copyright (c) 2022, Linaro. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

# Exit on command failure
set -o errexit

# Fail on unset variable
# Use "${VARNAME-}" instead of "$VARNAME" to access unset variable(s)
set -o nounset

# Enable debug mode if $TRACE is set
# To enable, run with: "env TRACE=1 ./new-device.sh"
if [[ "${TRACE-0}" == "1" ]]; then
    set -o xtrace
fi

# Check for too many parameters
if [ $# -gt 1 ]
  then
    echo "Too many paramters provided."
	echo "Run './new-device.sh -h' for help."
	exit 1
fi

# Check if the first arg is -h or --help
if [[ "${1-}" =~ ^-*h(elp)?$ ]]; then
    echo "Usage: ./new-device.sh [hostname]

Simulates provisioning of a new client device.

This script generates a random UUID to identify a client hardware device, then
creates a certificate signing request (CSR) which includes that UUID. It then
calls 'make_csr_cbor.go', which converts the CSR into a CBOR file that can be
sent to the CA in the bootstrap server using the REST API. The encoded CSR file
is then sent to the CA server for processing, which will return a DER-encoded
certificate in a CBOR wrapper. This DER-encoded CBOR-wrapped certificate is
finally converted to a PEM-encoded certificate for convenience sake.

The following files are placed them in the 'certs' folder:

- certs/{UUID}.crt    PEM-encoded client device certificate
- certs/{UUID}.key    The private key associated with {UUID}.crt

You can view the content of the certificate via:

   $ openssl x509 -in certs/{UUID}.crt -noout -text

HOSTNAME
--------
A consistent hostname must be used in your network layout, since the name will
be included in the generated device certificate, and the TLS handshake will
fail if the hostname used by the device certificate and the server certificate
don't match.

For this script, which generates a device certificate that includes the
hostname in the ORG field of the certificate subject, you can set the hostname
value through several mechanism:

   1. Via the '[hostname]' parameter when calling this script, i.e.:

      $ ./new-device.sh myhostname.local

   2. Setting the 'CAHOSTNAME' environment variable before running this script:

      $ export CAHOSTNAME=myhostname.local
      $ ./new-device.sh

   3. Not doing anything, which will cause the script to evaluate the system
      hostname via the 'hostname' command.

NOTE: 'localhost' is useful for testing, particularly if you are behind a NAT,
but won't allow access from a remote device. In order for this server to work
in that network topology, you'll need to set the hostname to a valid DNS name
that resolves to this host.

If you get an error like 'failed: Connection refused', make sure that you are
setting the correct hostname value before running this script.
"
    exit
fi

# These steps can be followed to simulate a certificate request from a
# hardware device.
#
# This test scenario calls `make_csr_cbor.go`, which takes a
# certificate signing request (CSR) file, and converts it into a CBOR
# file that can be sent to the CA server using the REST API.  The
# encoded CSR file can then be sent to the CA server using `wget`,
# which will return the generated certificate as a CBOR payload.

# Please follow the steps in `README.md` and make sure the liteboot
# server is running (`run-server.sh`), before running this script.

# Resolve hostname
if [ $# -eq 1 ]
  then
    # Use command line parameter for hostname value if present
    HOSTNAME=$1
  else
    # Check for CAHOSTNAME env variable, default to 'hostname' cmd if undefined
    HOSTNAME=${CAHOSTNAME:-$(hostname)}
fi

# Generate a device ID.  BSD's uuidgen outputs uppercase, so conver
# that here.
DEVID=$(uuidgen | tr '[:upper:]' '[:lower:]')
DEVPATH=certs/$DEVID
DEVVENDOR="Test Vendor"

echo New device: "$DEVID"

# Generate a private user key for this device.
openssl ecparam -name prime256v1 -genkey -out "$DEVPATH.key"

# Generate the CSR for this key.
openssl req -new \
	-key "$DEVPATH.key" \
	-out "$DEVPATH.csr" \
	-subj "/O=$DEVVENDOR/CN=$DEVID/OU=LinaroCA Device Cert - Signing"

# Convert this CSR to cbor.
go run make_csr_cbor.go -in "$DEVPATH.csr" -out "$DEVPATH.cbor"

# Submit the CSR.
wget --ca-certificate=certs/SERVER.crt \
	--certificate=certs/BOOTSTRAP.crt \
	--private-key=certs/BOOTSTRAP.key \
	--post-file "$DEVPATH.cbor" \
	--header "Content-Type: application/cbor" \
	"https://$HOSTNAME:1443/api/v1/cr" \
	-O "$DEVPATH.rsp"

# When this is successfully processed by the CA, it will return a DER
# encoded certificate enclosed in a CBOR wrapper.  The following
# commands will convert this to a PEM-encoded certificate file.
go run get_cert_cbor.go -in "$DEVPATH.rsp" -out "$DEVPATH.crt"

# Display the certificate
openssl x509 -in "$DEVPATH.crt" -noout -text

# Verify the generated certificate against the CA.
openssl verify -CAfile certs/CA.crt "$DEVPATH.crt"

# Delete the files that aren't needed
rm "$DEVPATH.csr"
rm "$DEVPATH.cbor"
rm "$DEVPATH.rsp"
