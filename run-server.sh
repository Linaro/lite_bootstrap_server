#!/bin/bash

HOSTNAME=localhost

if [ ! -f CA.crt -o ! -f CA.key -o ! CA.crt ]; then
	echo "Server certificates not present.  Please run ./setup-ca.sh"
	exit 1
fi

# Build linaroca
go build || exit 1

# Start linaroca
./linaroca server start -p 1443
