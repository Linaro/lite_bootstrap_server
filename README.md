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
        -subj "/O=Linaro, LTD/CN=LinaroCA HTTP Server Cert"
```

The contents of the certificate can be verified via:

```bash
$ openssl x509 -in SERVER.crt -noout -text
Certificate:
    Data:
        Version: 1 (0x0)
        Serial Number: 9242433173333109773 (0x8043b800acce6c0d)
    Signature Algorithm: ecdsa-with-SHA256
        Issuer: O=Linaro, LTD, CN=LinaroCA HTTP Server Cert
        Validity
            Not Before: Jun 25 11:58:29 2020 GMT
            Not After : Jun 23 11:58:29 2030 GMT
        Subject: O=Linaro, LTD, CN=LinaroCA HTTP Server Cert
        Subject Public Key Info:
            Public Key Algorithm: id-ecPublicKey
                Public-Key: (256 bit)
                pub: 
                    04:8b:9e:75:7b:92:1a:4c:c9:86:73:f0:4d:a1:04:
                    41:17:81:8e:79:a5:5a:ce:df:b9:81:28:a4:43:49:
                    20:f2:e6:af:54:77:dc:44:23:4f:d2:80:96:a3:aa:
                    b1:c5:d4:7f:be:cb:1a:d1:d1:b7:a5:4b:d7:8d:19:
                    28:75:19:0d:e4
                ASN1 OID: prime256v1
                NIST CURVE: P-256
    Signature Algorithm: ecdsa-with-SHA256
         30:45:02:20:14:e4:dc:44:68:9b:47:4b:ef:3c:fe:f6:95:bd:
         22:d8:89:ab:da:74:67:89:b3:b3:b1:52:91:55:98:fc:ca:76:
         02:21:00:d0:98:eb:5d:b7:c1:62:fe:80:e8:fa:bd:64:f5:d7:
         51:28:a2:18:42:8f:ec:f9:97:fd:18:9e:e1:71:9a:b4:ac
```