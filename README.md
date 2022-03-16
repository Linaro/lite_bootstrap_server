# Linaro Certificate Authority

A basic, proof-of-concept certificate authority (CA) and utility written in Go.

This utility can be used to validate and sign certificate signing requests
(CSRs), list previously registered devices, and enables a basic HTTP server
and REST API that can be used to communicate with the CA and it's underlying
device registry.

Certificate requests are logged to a local SQLite database (`CADB.db`), which
can be used to extend the utility for blocklist/allowlist type features,
and certificate status checks based on the registered device's unique ID.

## Quick Setup

### Setup Script

Three bash scripts are provided:

- `setup-ca.sh`
- `run-server.sh`
- `new-device.sh`

The first script creates the HTTP and CA keys and certificates.  This
should only need to be run once, and, in fact, will require existing
certs to have to be removed manually before it can be run again.

The second script will start the CA server on port 1443.

The last script creates a new device, and registers it with this CA.
The certificates for the device will be placed in the 'certs'
directory, along with the certificates above.

For details, or to perform these steps manually, please refer directly
to these scripts.

## Testing mutual TLS Authentication

A secondary TCP server is started up along with the main CA server to test
mutual TLS authentication using client certificates.

mutual TLS authentication requests a certificate from the connecting client
device that has been signed with the CA, adding an additional level of trust
on behalf of the server concerning the client device.

### Using a CA-signed client certificate

Once a user certificate has been generated (see steps above), you can test
mTLS connections via:

```bash
$ openssl s_client \
  -cert certs/USER.crt -key certs/USER.key \
  -CAfile SERVER.crt \
  -connect localhost:8443
```

- The `cert` and `key` fields indicates the client certificate and key
- The `CAfile` field is the server certificate, signed by the CA

The CA key is not added in this situation since the TCP server has a
copy of it that it uses to validate the client certificate's signature
against.

If the connection is successful, you should get the following response
at the end of the outptut:

```
Verify return code: 0 (ok)
```

And linaroca will display details about the client cert in the log output:

```bash
$ ./linaroca server start
Starting CA server on port https://localhost:1443
Starting mTLS TCP server on localhost:8443
Connection accepted from 127.0.0.1:50231
Client certificate:
- Issuer CN: LinaroCA Root Cert - 2020
- Subject: CN=LinaroCA Device Cert - Signing,
           OU=cb0dbc8a-2030-4799-a7af-183fddff04d7,O=localhost
```

### Using an invalid client certificate

To test with an **invalid user certificate**, generate a new cert:

```bash
$ openssl ecparam -name prime256v1 -genkey -out USERBAD.key
$ openssl req -new -x509 -sha256 -days 365 -key USERBAD.key -out USERBAD.crt -subj "/O=Linaro, LTD/CN=localhost"
```

Then try the request again:

```bash
$ openssl s_client \
  -cert USERBAD.crt -key USERBAD.key \
  -CAfile certs/SERVER.crt\
  -connect localhost:8443
```

You should get the following error from linaroca:

```
Connection accepted from 127.0.0.1:50443
Client handshake error: tls: failed to verify client certificate: x509: certificate signed by unknown authority
```

# REST API Endpoints

The REST API is a **work in progress**, and not fully implemented at present!

API based loosely on [CMP (RFC4210)](https://tools.ietf.org/html/rfc4210).

## `/api/v1/cr` Certification Request

This API requires the Content-Type to be set on the post data, it will
return an error if given the default Content-Type, and it should be
set to either `application/cbor` or `application/json`.  The request
is identical, and merely differs in how the data is encoded.  A
request made in cbor will have a cbor reply, as will json.

The content type can be specified with wget by adding:

```
--header "Content-Type: application/cbor"
```

as an argument to wget.  The content type must match the encoding of
the posted file.

### `/api/v1/cr` Certification Request from CBOR: **POST**

This endpoint is used to request a certificate for a new device,
providing a CSR payload wrapped in a single CBOR array:

```cddl
[ bstr ]
```

> You can generate the expected CBOR payload from a CSR file via the
> `make_csr_cbor.go` helper described elsewhere in this document.

```cbor
[
  b64'MIH9MIGlAgEAMEMxEjAQBgNVBAoMCWxvY2FsaG9zdDEtMCsGA1UEAwwkZThjNDdiNDItZjdmYy00ZGM4LWI1MzgtOTM0OTZiNjE5YTNjMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEJ5aLGLJID8MVHCHEzQOlO63RvBaTy8lRbtkDODYPgDKBOuAHWXbytgjO32K8O282BK/Hl5eEKqcXcHerlxE2xKAAMAoGCCqGSM49BAMCA0cAMEQCIEFoH+NV9jXJA0PmctbCQ7FOBnE/aT9hmqBuvBq2kIhuAiAyKGAUIAzHDBZ+lY6WaJGh/56rzr4KprVtNYFGLHLs1g=='
]
```

The CA will assign and record a unique serial number for this certificate,
which can later be used to check the certificate status via the `cs/{serial}`
endpoint.

It will reply with a CBOR result in the following format:

```cddl
{
   1 => int,   ; Status.
   2 => bstr,  ; Certificate
}
```

For example:
```cbor
{
  0: 0,
  1: b64'MIIBtjCCAVugAwIBAgIIFrKA6WV+D5gwCgYIKoZIzj0EAwIwOjEUMBIGA1UEChMLTGluYXJvLCBMVEQxIjAgBgNVBAMTGUxpbmFyb0NBIFJvb3QgQ2VydCAtIDIwMjAwHhcNMjExMDI5MTI0MjM0WhcNMjIxMDI5MTI0MjM0WjBDMRIwEAYDVQQKEwlsb2NhbGhvc3QxLTArBgNVBAMTJGU4YzQ3YjQyLWY3ZmMtNGRjOC1iNTM4LTkzNDk2YjYxOWEzYzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABCeWixiySA/DFRwhxM0DpTut0bwWk8vJUW7ZAzg2D4AygTrgB1l28rYIzt9ivDtvNgSvx5eXhCqnF3B3q5cRNsSjQjBAMB0GA1UdDgQWBBSCv0wCNUXMXavfi15AbcCclDvfkzAfBgNVHSMEGDAWgBSxhUrvHyyKgHn5/FaoKd761df1tjAKBggqhkjOPQQDAgNJADBGAiEAjDVYvr1qBfvc0VFZcFLxwO/5XvnBh2jZFpL9ykKsCw8CIQDF3ne7yokRAHt0nn35CW/J3FclGYH9rBVCZr7FU+pzHg==',
}
```

### `/api/v1/cr` Certification Request from JSON: **POST**

This endpoint is used to request a certificate for a new device, providing a
**BASE64-encoded CSR payload in a JSON wrapper** with the following format:

> You can generate the expected JSON payload from a CSR file via the
  `make_csr_json.go` helper described elsewhere in this document.

```json
{
  "CSR":"MIH9MIGlAgEAMEMxEjAQBgNVBAoMCWxvY2FsaG9zdDEtMCsGA1UEAwwkZThjNDdiNDItZjdmYy00ZGM4LWI1MzgtOTM0OTZiNjE5YTNjMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEJ5aLGLJID8MVHCHEzQOlO63RvBaTy8lRbtkDODYPgDKBOuAHWXbytgjO32K8O282BK/Hl5eEKqcXcHerlxE2xKAAMAoGCCqGSM49BAMCA0cAMEQCIEFoH+NV9jXJA0PmctbCQ7FOBnE/aT9hmqBuvBq2kIhuAiAyKGAUIAzHDBZ+lY6WaJGh/56rzr4KprVtNYFGLHLs1g=="
}
```

> :warning: The `CSR` field in the JSON payload is the BASE64 encoded byte
  array from the equivalent of a DER file, and doesn't include the text
  headers present in the PEM file generated by
  `openssl req -new -key USER.key -out USER.csr` lower in this guide.
  `make_csr_json.go` takes care of the process of loading the human-readable
  PEM file, extracting the byte array for the CSR, and encoding the payload
  in the expected format, seen below.

The CA will assign and record a unique serial number for this certificate,
which can later be used to check the certificate status via the `cs/{serial}`
endpoint.

It will reply with a JSON array containing `Status` and `Cert` fields,
where `Cert` contains the BASE64-encoded certificate:

```json
{
  "Status":0,
  "Cert":"MIIBtjCCAVugAwIBAgIIFrKA6WV+D5gwCgYIKoZIzj0EAwIwOjEUMBIGA1UEChMLTGluYXJvLCBMVEQxIjAgBgNVBAMTGUxpbmFyb0NBIFJvb3QgQ2VydCAtIDIwMjAwHhcNMjExMDI5MTI0MjM0WhcNMjIxMDI5MTI0MjM0WjBDMRIwEAYDVQQKEwlsb2NhbGhvc3QxLTArBgNVBAMTJGU4YzQ3YjQyLWY3ZmMtNGRjOC1iNTM4LTkzNDk2YjYxOWEzYzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABCeWixiySA/DFRwhxM0DpTut0bwWk8vJUW7ZAzg2D4AygTrgB1l28rYIzt9ivDtvNgSvx5eXhCqnF3B3q5cRNsSjQjBAMB0GA1UdDgQWBBSCv0wCNUXMXavfi15AbcCclDvfkzAfBgNVHSMEGDAWgBSxhUrvHyyKgHn5/FaoKd761df1tjAKBggqhkjOPQQDAgNJADBGAiEAjDVYvr1qBfvc0VFZcFLxwO/5XvnBh2jZFpL9ykKsCw8CIQDF3ne7yokRAHt0nn35CW/J3FclGYH9rBVCZr7FU+pzHg=="
}
```

## `/api/v1/p10cr` Certification Request from PKCS10: **POST**

This endpoint is used to request a certificate for a new device, posting a
PKCS#10 CSR file in **PEM format** (ex. `USER.csr` generated lower in this
guide) as part of the request.

The CA will assign and record a unique serial number for this certificate,
which can later be used to check the certificate status via the `cs/{serial}`
endpoint.

It will reply with a certificate file in PEM format or a status error in JSON,
depending on the input CSR provided.

Testing this endpoint with:

```bash
$ curl -v --cacert SERVER.crt \
  -F csrfile=@USER.csr \
  --output USER.crt \
  https://localhost:1443/api/v1/p10cr
```

should give you a `USER.crt` file in **PEM format**, which you can view via:

```bash
$ openssl x509 -in USER.crt -noout -text
```

## `api/v1/cs/{serial}` Certificate Status Request: **GET**

Requests the certificate status based on the supplied certificate serial number.

The serial number is generated by the CA during the certificate generation
process (`/ap1/v1/cr`), and is a unique timestamp-based 64-bit integer (ex. 
`1635511354607407000`) that is added to the certificate before sending it back
to the requesting device.

It can be retrieved from a certificate via the serial number field, for
example:

```bash
$ openssl x509 -in USER.crt -noout -text
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number: 1635511354607407000 (0x16b280e9657e0f98)
        ...
```

```bash
$ curl -v --cacert SERVER.crt \
  https://localhost:1443/api/v1/cs/1635511354607407000
```

This endpoint will return one of three JSON response types:

- `{"status": "1"}` + HTTP response code **200**: Indicating that the serial
  number exists, and that the certificate is marked as **valid** in the CA
  database.
- `{"status": "0"}` + HTTP response code **200**: Indicating that the serial
  number exists, and that the certificate is marked as **invalid** in the
  CA database (i.e., it has been **revoked**).
- `{"error": "<error msg>"}` + HTTP response code **400**: Indicating one
  of the following error messages:
  - `invalid request`: Poorly formatted serial number was provided
  - `invalid serial number`: No certificate matching supplied serial found

## `api/v1/kur` Key Update Request: **POST**

Request an update to an existing (non-revoked and non-expired) certificate. An
update is a replacement certificate containing either a new subject public
key or the current subject public key.

## `api/v1/krr` Key Revocation Request: **POST**

Requests the revocation of an existing certificate registration.
