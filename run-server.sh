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

# Check if the first arg is -h or --help
if [[ "${1-}" =~ ^-*h(elp)?$ ]]; then
    echo 'Usage: ./run-server.sh

Starts the bootstrap server and CA.

To set the hostname used by the server, run the following command before
executing this script, replacing "localhost" with an appropriate hostname:

   $ export CAHOSTNAME=localhost
'
    exit
fi

if [ ! -f certs/CA.crt -o ! -f certs/CA.key ]; then
	echo "Server certificates not present.  Please run ./setup-ca.sh"
	exit 1
fi

# Build liteboot
go build -o liteboot || exit 1

# Run the server, listening by default on port 1443.
./liteboot server start -p 1443

# This will serve web pages from root, and handle REST API requests
# from the `/api/v1` sub-path, with page routing handled in
# `httpserver.go`.
#
# Note: A secondary TCP server is started at the same time to test
# mutual TLS (mTLS) connections.  This can be ignored at this point.
