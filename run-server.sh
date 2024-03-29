#!/usr/bin/env bash
# Copyright (c) 2022, Linaro. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

# Exit on command failure
set -o errexit

# Fail on unset variable
# Use "${VARNAME-}" instead of "$VARNAME" to access unset variable(s)
set -o nounset

# Enable debug mode if $TRACE is set
# To enable, run with: "env TRACE=1 ./run-server.sh"
if [[ "${TRACE-0}" == "1" ]]; then
    set -o xtrace
fi

# Check for too many parameters
if [ $# -gt 1 ]
  then
    echo "Too many paramters provided."
	echo "Run './run-server.sh -h' for help."
	exit 1
fi

# Check if the first arg is -h or --help
if [[ "${1-}" =~ ^-*h(elp)?$ ]]; then
    echo "Usage: ./run-server.sh [hostname]

Starts the bootstrap server and CA.

HOSTNAME
--------
If you wish to use a specific HOSTNAME for the servers, set the correct value
before running this script via one of:

   1. Setting the optional [hostname] parameter in this script:

      $ ./run-server.sh myhostname.local

   2. Adding 'hostname = \"myhostname.local\"' to .liteboot.toml

   3. Setting 'CAHOSTNAME' before executing this script:

      $ export CAHOSTNAME=myhostname.local
      $ ./run-server.sh

NOTE: 'localhost' is useful for testing, particularly if you are behind a NAT,
but won't allow access from a remote device. In order for this server to work
in that network topology, you'll need to set the hostname to a valid DNS name
that resolves to this host.

This hostname must be used consistently in your network layout, since the name
will be included in the generated certificates, and the TLS handshake will fail
if the hostname used by the servers and the value defined in the certificate(s)
don't match.

If you get an error like 'failed: Connection refused', make sure that you are
setting the correct hostname value before running this script.
"
    exit
fi

if [ ! -f certs/CA.crt ] || [ ! -f certs/CA.key ]; then
	echo "Server certificates not present.  Please run ./setup-ca.sh"
	exit 1
fi

# Build liteboot if necessary
go build -o liteboot || exit 1

# Run the server, listening by default on port 1443.
if [ $# -eq 1 ]
  then
    # Use command line parameter for hostname value
    ./liteboot server start -p 1443 --hostname="$1"
  else
    # Let liteboot resolve hostname on it's own
    ./liteboot server start -p 1443
fi

# This will serve web pages from root, and handle REST API requests
# from the `/api/v1` sub-path, with page routing handled in
# `httpserver.go`.
#
# Note: A secondary TCP server is started at the same time to test
# mutual TLS (mTLS) connections.  This can be ignored at this point.
