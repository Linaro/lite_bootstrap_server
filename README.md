# Linaro Certificate Authority

A basic, proof-of-concept certificate authority (CA) and utility written in Go.

This utility can be used to validate and sign certificate signing requests (CSRs),
list previously registered devices, and enables a basic HTTP server that can be
used to communicate with the CA and it's underlying device registry.

## Quick Setup 

### Key Generation for HTTP Server

The HTTP server requires a private key for TLS, which can be generated via:

```bash
$ openssl ecparam -name secp256r1 -genkey -out SERVER.key
```

You can then generate a self-signed X.509 public key via:

```bash
$ openssl req -new -x509 -sha256 -days 3650 -key SERVER.key -out SERVER.crt \
        -subj "/O=Linaro, LTD/CN=LinaroCA HTTP Server Cert"
```

This public key should be available on any devices connecting to the CA to
verify that we are communicating with the intended CA server.

The contents of the certificate can be verified via:

```bash
$ openssl x509 -in SERVER.crt -noout -text
```

## HTTPS Server

To initialise the HTTPS server on the default port (443), run:

> :information_source: Use the `-p <port>` flag to change the port number.

```bash 
$ linaroca server start
Starting HTTPS server on port https://localhost:443
```

This will serve web pages from root, and handle REST API requests from the
`/api/v1` sub-path.

### REST API Endpoints

API based loosely on [CMP (RFC4210)](https://tools.ietf.org/html/rfc4210).

#### Initialisation Request: `/api/v1/ir` **POST**

This endpoint is used to register new (previously unregistered) devices into
the management system. Initialisation must occur before certificates can be
requested.

A unique serial number must be provided for the device, and any certificates
issued for this device will be associated with this device serial number.

#### Certification Request: `/api/v1/cr` **POST**

This endpoint is used for certificate requests from existing devices who
wish to obtain new certificates.

The CA will assign and record a unique serial number for this certificate,
which can be used later to check the certificate status via the `cs` endpoint.

#### Certification Request from PKCS10: `/api/v1/p10cr` **POST**

This endpoint is used for certificate requests from existing devices who
wish to obtain new certificates, providing a PKCS#10 CSR file for the request.

The CA will assign and record a unique serial number for this certificate,
which can be used later to check the certificate status via the `cs` endpoint.

#### Certificate Status Request: `api/v1/cs/{serial}` **GET**

Requests the certificate status based on the supplied certificate serial number.

#### Key Update Request: `api/v1/kur` **POST**

Request an update to an existing (non-revoked and non-expired) certificate. An
update is a replacement certificate containing either a new subject public
key or the current subject public key.

#### Key Revocation Request: `api/v1/krr` **POST**

Requests the revocation of an existing certificate registration.
