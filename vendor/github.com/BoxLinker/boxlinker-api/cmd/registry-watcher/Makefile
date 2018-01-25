
registry:
	docker rm -f boxlinker-registry-watcher || true
	docker run -p 5000:5000 -p 5001:5001 --name boxlinker-registry-watcher -v `pwd`/registry_config.yml:/etc/docker/registry/config.yml registry:2

# openssl genrsa -out ./private_key.pem 4096
# openssl req -new -x509 -key ./private_key.pem -out ./root.crt -days 3650 -subj /C=CN/ST=state/L=CN/O=cloverstd/OU=cloverstd\ unit/CN=boxlinker.com/emailAddress=service@boxlinker.com
