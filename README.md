# LITE Bootstrap Server

A proof-of-concept certification authority (CA) and bootstrap server written
in Go.

This utility provides an authenticated REST API that can be used to:

- Validate and process certificate signing requests (CSRs)
- Register devices on a 3rd party IoT device management service during certificate registration
- Provision devices with basic connection details (e.g. MQTT broker domain/port)
- List previously registered devices
- Check certificate validity

Certificate signing requests are logged to a local SQLite database (`CADB.db`).

Device registration hooks to an IoT device management service and returns connection
details. Currently, the returned connection details are MQTT broker details
based on Azure IoT Hub. Support for other providers may be offered in the future.

It also enables a basic TCP/TLS server (`mTLS TCP server`) that requests a client
certificate during the TLS handshake, verifying that the certificate has been
signed by the CA. This is only used for testing and proof of concept purposes.

> The REST API will only accept requests from client devices with access to the
  bootstrap certificate and private key generated by `setup-bootstrap.sh`
  further below in this guide. It's assumed that this certificate and its 
  corresponding private key is provisioned to devices in the factory. 

## Prerequisites

This project requires a recent version of go to compile and debug the server.

For platform-specific details on installing go, see: https://go.dev/doc/install

# Quick Setup

## 0. Build the App

If you haven't already, the first step is to clone and build the app locally:

```bash
$ git clone https://github.com/Linaro/lite_bootstrap_server.git
$ cd lite_bootstrap_server
$ go build -o liteboot
```

> The `-o liteboot` is optional, but will cause the compiled app to have the
  shorter `liteboot` name instead of the cumbersome default of
  `lite_bootstrap_server`.

## 1. Add a Config File

You can override various config settings with a `.liteboot.toml` file in the
application root folder.

The `[server]` settings, which map to the command-line parameters shown with
`liteboot server --help`, must be populated to set the ports used by the
utility, as well as to provide a response to the `ccs` endpoint:

```toml
[server]
# Azure IoT Hub Settings
hubname = "azure-hub-name"
resourcegroup = "azure-resource-group"
mqttport = 8883

# Set the server hostname explicitly
# hostname = "myhostname.local"

# CA port number (REST API)
port = 1443

# mTLS port number
mport = 8443
```

## 2. Set the Hostname

The hostname is important since it will be included in the subject line of
certificates signed by liteboot, and be part of the SERVER certificate used to
establish a TLS connection to the server.

The **same hostname** must be used in the SERVER certificate or the
TLS connection will be refused.

This application will attempt to determine the hostname to use for the
server(s) based on the following order of precedence:

1. Via the `--hostname` parameter in the `server` command
2. Via a `hostname` entry in the `.liteboot.toml` config file
   (see step 1 above)
3. Check for the `$CAHOSTNAME` environment variable
4. Check the system hostname (same value as `hostname` in shell)
5. Default to `localhost` as a last resort

For example, set the hostname before starting the server via:

```bash
$ export CAHOSTNAME='myhostname.local'
````

Which could also be overridden via:

```bash
$ ./liteboot server start --hostname="localhost"
```

For local development, this can usually be set to `localhost`, but will need
to be changed when working with multiple machines or containers.

## 3. Run Setup Scripts

First, create the HTTP and CA keys and certificates via `setup-ca.sh`.

This should only need to be **run once**, and, in fact, will require existing
certs to have to be removed manually before it can be run again:

> If you haven't set the hostname via the `CAHOSTNAME` environment variable,
  and wish to use a hostname other than what the `hostname` command resolves
  to, this script should be called with the hostname as the first parameter.
  Run the script with `-h` for details.

```bash
$ ./setup-ca.sh
```

Next, setup the bootstrap key pair via `setup-bootstrap.sh`.

This key pair is required to authenticate with the CA server, and the data
placed in `bootstrap_crt.txt` and `bootstrap_key.txt` will need to be
included in the client device, or included in any attempts to connect to the
REST API using `curl`, `wget`, etc.

The script can be run multiple times (after removing the generated
files), but only would need to be rerun after regenerating the CA
cert.

```bash
$ ./setup-bootstrap.sh
```

## 4. Start the Server

Run `run-server.sh` to start the CA server on port 1443, or whatever port you
set in the config file:

> This script also accepts an optional hostname override as the first
  parameter. Run the script with `-h` for details.

```bash
$ ./run-server.sh
```

## 5. Optional Steps

### Generate Test Device(s)

You can optionally run `new-device.sh` to create a new device, and registers
it with this CA.

The certificates for the device will be placed in the 'certs'
directory, along with the certificates above.

> This certificate is only useful for testing the CA, and generally will not
  be used in a production setting.

> This script also accepts an optional hostname override as the first
  parameter. Run the script with `-h` for details.

```bash
$ ./new-device.sh
```

### Cleanup

You can remove all existing certificate artifacts via:

```bash
$ rm CADB.rb
$ rm -rf certs
```

# REST API

The REST API is a **work in progress**, and may be changed in the future!

## `/api/v1/cr` Certification Request: **POST**

Request a certificate for a new device, based on the provided certificate
signing request (CSR).

The CA will assign and record a unique serial number for this certificate,
which can later be used to check the certificate status via the `cs/{serial}`
endpoint.

> This API requires the `Content-Type` to be set on the post data, and must be
  set to either `application/cbor` or `application/json`.  A request made in
  cbor will have a cbor reply, json will reply with a json encoded response.

### Request with `application/cbor`

The CSR payload should be wrapped in a single CBOR array:

```cddl
[ bstr ]
```

#### Example

See the `new-device.sh` script for an example of generating an appropriate
request from the shell, and saving the response.

#### Response

Replies with a CBOR array containing the following fields:

```cddl
{
   1 => int,   ; Status.
   2 => bstr,  ; Certificate
}
```

- `Status` is an error code where `0` indicates success, and non-zero values
should be treated as an error.
- `Certificate` contains the BASE64-encoded DER format certificate.

### Request with `application/json`

Takes a **BASE64-encoded CSR payload in a JSON wrapper** as follows:

```json
{
  "CSR":"MIH9MIGlAgEAMEMxEjAQBgNVBAoMCWxvY2FsaG9zdDEtMCsGA1UEAwwkZThjNDdiNDItZjdmYy00ZGM4LWI1MzgtOTM0OTZiNjE5YTNjMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEJ5aLGLJID8MVHCHEzQOlO63RvBaTy8lRbtkDODYPgDKBOuAHWXbytgjO32K8O282BK/Hl5eEKqcXcHerlxE2xKAAMAoGCCqGSM49BAMCA0cAMEQCIEFoH+NV9jXJA0PmctbCQ7FOBnE/aT9hmqBuvBq2kIhuAiAyKGAUIAzHDBZ+lY6WaJGh/56rzr4KprVtNYFGLHLs1g=="
}
```

> :warning: The `CSR` field in the JSON request payload is a BASE64 encoded
  byte array from a **DER file**. This paylaod doesn't include the text
  headers present in the PEM files generated by
  `openssl req -new -key UUID.key -out UUID.csr` in `new-device.sh`.

#### Response

Replies with a JSON array containing `Status`, and `Cert` fields:

- `Status` is an error code where `0` indicates success, and non-zero values
should be treated as an error.
- `Cert` contains the BASE64-encoded DER format certificate.

```json
{
  "Status":0,
  "Cert":"MIIBtjCCAVugAwIBAgIIFrKA6WV+D5gwCgYIKoZIzj0EAwIwOjEUMBIGA1UEChMLTGluYXJvLCBMVEQxIjAgBgNVBAMTGUxpbmFyb0NBIFJvb3QgQ2VydCAtIDIwMjAwHhcNMjExMDI5MTI0MjM0WhcNMjIxMDI5MTI0MjM0WjBDMRIwEAYDVQQKEwlsb2NhbGhvc3QxLTArBgNVBAMTJGU4YzQ3YjQyLWY3ZmMtNGRjOC1iNTM4LTkzNDk2YjYxOWEzYzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABCeWixiySA/DFRwhxM0DpTut0bwWk8vJUW7ZAzg2D4AygTrgB1l28rYIzt9ivDtvNgSvx5eXhCqnF3B3q5cRNsSjQjBAMB0GA1UdDgQWBBSCv0wCNUXMXavfi15AbcCclDvfkzAfBgNVHSMEGDAWgBSxhUrvHyyKgHn5/FaoKd761df1tjAKBggqhkjOPQQDAgNJADBGAiEAjDVYvr1qBfvc0VFZcFLxwO/5XvnBh2jZFpL9ykKsCw8CIQDF3ne7yokRAHt0nn35CW/J3FclGYH9rBVCZr7FU+pzHg==",
}
```

## `/api/v1/p10cr` Certification Request from PKCS10: **POST**

This endpoint is used to request a certificate for a new device, posting a
PKCS#10 CSR file in **PEM format**.

The CA will assign and record a unique serial number for this certificate,
which can later be used to check the certificate status via the `cs/{serial}`
endpoint.

It will reply with a certificate file in PEM format or a status error in JSON,
depending on the input CSR provided.

Testing this endpoint with:

> The example belows assumes `MBP2021.lan` as the local hostname. This
  value will vary from one machine to another.

```bash
$ curl -v --cacert certs/CA.crt  \
          --cert certs/BOOTSTRAP.crt \
          --key certs/BOOTSTRAP.key  \
          -F csrfile=@USER.csr       \
          --output USER.crt          \
          https://MBP2021.lan:1443/api/v1/p10cr
```

should give you a `USER.crt` file in **PEM format**, which you can view via:

```bash
$ openssl x509 -in USER.crt -noout -text
```

## `api/v1/ds/{uuid}` Device Status Request: **GET**

Checks if any valid certificates are associated with the specified device UUID.
If any valid certificates are found, their serial number will be returned in
the response payload.

This endpoint accepts requests in cbor (`application/cbor`) or json
(`application/json`), defaulting to JSON responses if no Content-Type is
provided. The response Content-Type will match the request type used.

### Request with `application/json`

```bash
$ curl -v --cacert certs/CA.crt  \
          --cert certs/BOOTSTRAP.crt \
          --key certs/BOOTSTRAP.key  \
          https://MBP2021.lan:1443/api/v1/ds/56d38f73-3f6f-4a59-86bc-d315a1ccc634
```

### Request with `application/cbor`

You can send a status request via:

```bash
$ curl -v --cacert certs/CA.crt  \
          --cert certs/BOOTSTRAP.crt \
          --key certs/BOOTSTRAP.key  \
          -H 'Content-Type: application/cbor' \
          https://MBP2021.lan:1443/api/v1/ds/56d38f73-3f6f-4a59-86bc-d315a1ccc634 \
          -o cbor_enc.raw
```

> CBOR requests receive binary responses, which curl will print as an unknown 
  symbol on the console. Pass the `-o <filename>` argument to store the binary 
  output in a file, and execute `go run cbor_decoder.go -i cbor_enc.raw -r ds`
  to decode  the binary output response and print in the JSON format.

### Responses

This endpoint will return the following responses, in JSON or CBOR encoding
depening on the `Content-Type` used:

- HTTP response code **200**
  - `{"Status":0,"Serials":null}`: No valid certs found for UUID
  - `{"Status":1,"Serials":[1649073924922206000]}`: Valid cert(s) found for UUID
- HTTP response code **400** + `{"error": "<error msg>"}`, where error msg is:
  - `need a valid UUID`: Invalid or improperly formatted UUID was provided
  - `unable to query db for UUID`: error querying for database UUID (fatal)
  - `bad request: Content-Type must be ...`: Invalid Content-Type provided

## `api/v1/cs/{serial}` Certificate Status Request: **GET**

Requests the certificate status based on the supplied certificate serial number.

The serial number is generated by the CA during the certificate generation
process (`/ap1/v1/cr`), and is a unique timestamp-based 64-bit integer (ex.
`1635511354607407000`) that is added to the certificate before sending it back
to the requesting device.

It can be retrieved using the `/api/v1/ds/{uuid}` endpoint, or from a
certificate file directly via the serial number field, for example:

> The example belows assumes a specific device UUID, and `MBP2021.lan` as
  the local hostname. These values will vary from one machine to another.

```bash
$ openssl x509 -in certs/a8c6f808-b659-4f88-affb-40498834c572.crt -noout -text
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number: 1648935985023194000 (0x16e2328abcaa9390)
    Signature Algorithm: ecdsa-with-SHA256
        Issuer: O=Linaro, LTD, CN=LRC - 20220330062226
        Validity
            Not Before: Apr  2 21:46:25 2022 GMT
            Not After : Apr  2 21:46:25 2023 GMT
        Subject: O=MBP2021.lan, OU=LinaroCA Device Cert - Signing, CN=a8c6f808-b659-4f88-affb-40498834c572
        ...
```

Once you have the serial number, you can send a status request via:

### Request with `application/json` (default)

You can send a status request via:

```bash
$ curl -v --cacert certs/CA.crt  \
          --cert certs/BOOTSTRAP.crt \
          --key certs/BOOTSTRAP.key  \
          https://MBP2021.lan:1443/api/v1/cs/1648935985023194000
```

### Request with `application/cbor`

You can send a status request via:

```bash
$ curl -v --cacert certs/CA.crt  \
          --cert certs/BOOTSTRAP.crt \
          --key certs/BOOTSTRAP.key  \
          -H 'Content-Type: application/cbor' \
          https://MBP2021.lan:1443/api/v1/cs/1648935985023194000 \
          -o cbor_enc.raw
```

> CBOR requests receive binary responses, which curl will print as an unknown 
  symbol on the console. Pass the `-o <filename>` argument to store the binary 
  output in a file, and execute `go run cbor_decoder.go -i cbor_enc.raw -r cs`
  to decode  the binary output response and print in the JSON format.

### Responses

This endpoint will return the following responses, in JSON or CBOR depending
on the `Content-Type` used:

- HTTP response code **200**
  - `{"status": "1"}`: Indicates that the serial number exists, and that the
    certificate is marked as **valid** in the CA database.
  - `{"status": "0"}`: Indicates that the serial number exists, but that the
    certificate is marked as **invalid** in the CA database (i.e., it has been
    **revoked**).
- HTTP response code **400** + `{"error": "<error msg>"}`, where error msg is:
  - `invalid request`: Poorly formatted serial number was provided
  - `invalid serial number`: No certificate matching supplied serial found
  - `bad request: Content-Type must be ...`: Invalid Content-Type provided

## `api/v1/cc/{serial}` X509 Certificate Copy Request: **GET**

Requests a copy of the x509 certificate associated with the supplied serial
number. If found, the certificate data is returned in **PEM format** (BASE64
ASCII).

> See the `cs` endpoint for details on retrieving a certificate serial number.

### Request with `application/json` (default)

You can send the x509 certificate copy request via:

```bash
$ curl -v --cacert certs/CA.crt  \
          --cert certs/BOOTSTRAP.crt \
          --key certs/BOOTSTRAP.key  \
          https://MBP2021.lan:1443/api/v1/cc/1664276589086394391
```

### Request with `application/cbor`

You can send the x509 certificate copy request via:

```bash
$ curl -v --cacert certs/CA.crt  \          
          --cert certs/BOOTSTRAP.crt \
          --key certs/BOOTSTRAP.key  \
          -H 'Content-Type: application/cbor' \
          https://MBP2021.lan:1443/api/v1/cc/1664276589086394391 \
          -o cbor_enc.raw
```

> CBOR requests receive binary responses, which curl will print as an unknown 
  symbol on the console. Pass the `-o <filename>` argument to store the binary 
  output in a file, and execute `go run cbor_decoder.go -i cbor_enc.raw -r cc`
  to decode  the binary output response and print in the JSON format.

### Response

Replies with a JSON or CBOR array containing `Status`, and `Cert` fields:

- `Status` is an error code where `0` indicates success, and non-zero values
should be treated as an error.
- `Cert` contains the string type x509 pem format certificate.

For example, if the `Content-Type` is set to the default `application/json`:

```json
{
  "Status":0,
  "Cert":"-----BEGIN CERTIFICATE-----\nMIIB6TCCAY+gAwIBAgIIFxiyvXR+SBcwCgYIKoZIzj0EAwIwNTEUMBIGA1UEChML\nTGluYXJvLCBMVEQxHTAbBgNVBAMTFExSQyAtIDIwMjIwOTI3MTEwMjE3MB4XDTIy\nMDkyNzExMDMwOVoXDTIzMDkyNzExMDMwOVowfDEiMCAGA1UEChMZYXJtLXZpcnR1\nYWwtbWFjaGluZS5sb2NhbDEnMCUGA1UECxMeTGluYXJvQ0EgRGV2aWNlIENlcnQg\nLSBTaWduaW5nMS0wKwYDVQQDEyRjNTJmN2Y4NS04ZmYwLTQzOGQtOGEwNi05Mzlj\nOWJlNTlhNGYwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARrMDR1HgkP/ET4yNqG\nM5km89r2DFAAKZDyL2lVg3C8sjNVrxoR73G7nhC8EyCFMwrIMK99O92tGFWOWqaM\nb2r9o0IwQDAdBgNVHQ4EFgQUDBH9jc9zvi3un83TFPI/tHhMF8YwHwYDVR0jBBgw\nFoAU+m+L2p3gV7JPexhVX6PFWMFy9cowCgYIKoZIzj0EAwIDSAAwRQIgM5zSdrUY\n7f0QG3cq+5MVL/rAb4k4Ok6vQIE63zRwQt8CIQCsQ7NQV9Sk8y3YU0R3HGUlOOMI\nKKmII4ZjIWrZs7Qr/A==\n-----END CERTIFICATE-----\n"
}
```

## `api/v1/ccs` Cloud Connectvity Settings: **GET**

Requests the cloud connectivity details for the MQTT broker associated with
your cloud infrastructure.

This endpoint allows device to retrieve details about the MQTT broker, etc.,
that they should connect to, and is usually retrieved during the provisioning
process, or when an alert is received by the client device that the cloud
connectivity settings have changed.

### Request with `application/cbor`

You can send a status request via:

```bash
$ curl -v --cacert certs/CA.crt  \
          --cert certs/BOOTSTRAP.crt \
          --key certs/BOOTSTRAP.key  \
          -H 'Content-Type: application/cbor' \
          -H 'Accept: application/cbor' \
          https://MBP2021.lan:1443/api/v1/ccs \
          -o cbor_enc.raw
```

> CBOR requests receive binary responses, which curl will print as an unknown 
  symbol on the console. Pass the `-o <filename>` argument to store the binary 
  output in a file, and execute `go run cbor_decoder.go -i cbor_enc.raw -r ccs`
  to decode  the binary output response and print in the JSON format.

### Request with `application/json` (default)

You can send a status request via:

```bash
$ curl -v --cacert certs/CA.crt  \
          --cert certs/BOOTSTRAP.crt \
          --key certs/BOOTSTRAP.key  \
          https://MBP2021.lan:1443/api/v1/ccs
```

### Response

Replies with a JSON array containing the following fields:

```json
{
  "Hubname":"azure_hubname",
  "Port":8883
}
```

- `Hubname` contains the Azure IoT Hub hubname string.
- `Port` contains the Azure IoT Hub MQTT port number.

## `api/v1/kur` Key Update Request: **POST** (TODO)

Request an update to an existing (non-revoked and non-expired) certificate. An
update is a replacement certificate containing either a new subject public
key or the current subject public key.

## `api/v1/krr` Key Revocation Request: **POST** (TODO)

Requests the revocation of an existing certificate registration.

# Mutual TLS Test Server

A secondary TCP server is started up along with the main CA server to test
mutual TLS authentication using client certificates.

mutual TLS authentication requests a certificate from the connecting client
device that has been signed with the CA, adding an additional level of trust
on behalf of the server concerning the client device.

### Using a CA-signed client certificate

Once a user certificate has been generated (via `new-device.sh`), you can test
mTLS connections via:

> The example belows assumes a specific device UUID, and `MBP2021.lan` as
  the local hostname. These values will vary from one machine to another.

```bash
$ openssl s_client \
  -cert certs/f269528d-ff66-4fb0-83d8-e449e0038010.crt \
  -key certs/f269528d-ff66-4fb0-83d8-e449e0038010.key \
  -CAfile certs/CA.crt \
  -connect MBP2021.lan:8443
```

- The `cert` and `key` fields indicate the client certificate and key
- The `CAfile` field is the CA certificate

If the connection is successful, you should get the following response
at the end of the outptut:

```
Verify return code: 0 (ok)
```

And liteboot will display details about the client cert in the log output:

```bash
$ ./liteboot server start
Starting mTLS TCP server on MBP2021.lan:8443
Starting CA server on port https://MBP2021.lan:1443
Connection accepted from 127.0.0.1:60510
[Certificate 0]
  - Subject: CN=f269528d-ff66-4fb0-83d8-e449e0038010,OU=Signing,O=Linaro\, LTD
  - Serial:  1671014018808547000
[Certificate 1]
  - Subject: CN=LRC - 20221129052020,O=Linaro\, LTD
  - Serial:  2022001129172020
```

- `Certificate 0` is the device certificate
- `Certificate 1` is the CA certificate used to sign certificate 0.

### Using an invalid client certificate

To test with an **invalid user certificate**, generate a new cert:

```bash
$ openssl ecparam -name prime256v1 -genkey -out USERBAD.key
$ openssl req -new -x509 -sha256 -days 365 \
  -key USERBAD.key -out USERBAD.crt \
  -subj "/O=Linaro, LTD/CN=MBP2021.lan"
```

Then try the request again:

```bash
$ openssl s_client \
  -cert USERBAD.crt -key USERBAD.key \
  -CAfile certs/CA.crt \
  -connect MBP2021.lan:8443
```

You should get the following error from liteboot, since the device certificate
has not been signed by our CA:

```
Connection accepted from 127.0.0.1:50443
Client handshake error: tls: failed to verify client certificate: x509: certificate signed by unknown authority
```

# Troubleshooting

## FAQs

### `setup-bootstrap.sh` causes `Can't load /home/user/.rnd into RNG`

Run `openssl rand -writerand .rnd` to generate the missing .rnd file.
