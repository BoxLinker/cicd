#!/bin/bash
openssl genrsa -out ./private_key.pem 4096
openssl req -new -x509 -key ./private_key.pem -out ./root.crt -days 3650 -subj /C=CN/ST=state/L=CN/O=cloverstd/OU=cloverstd\ unit/CN=boxlinker.com/emailAddress=service@boxlinker.com
