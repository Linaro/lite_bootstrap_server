#!/bin/bash

: ${HOSTNAME:=localhost}

if [ ! -f certs/CA.crt -o ! -f certs/CA.key ]; then
	echo "Server certificates not present.  Please run ./setup-ca.sh"
	exit 1
fi

# Build linaroca
go build || exit 1

# Run the server, listening by default on port 1443.
./linaroca server start -p 1443

# This will serve web pages from root, and handle REST API requests
# from the `/api/v1` sub-path, with page routing handled in
# `httpserver.go`.
#
# Note: A secondary TCP server is started at the same time to test
# mutual TLS (mTLS) connections.  This can be ignored at this point.
