# Linaro Certificate Authority

A basic, proof-of-concept certificate authority (CA) and utility written in Go.

This utility can be used to validate and sign certificate signing requests (CSRs),
list previously registered devices, and enables a basic HTTP server that can be
used to communicate with the CA and it's underlying device registry.

## Key Generation for HTTP Server

The HTTP server requires a private key for TLS, which can be generated via:

```bash
$ openssl ecparam -name secp256r1 -genkey -out SERVER.key
```

You can then generate a self-signed X.509 public key via:

```bash
$ openssl req -new -x509 -sha256 -days 3650 -key SERVER.key -out SERVER.crt \
        -subj "/O=Linaro/CN=CAWebServer"
```
