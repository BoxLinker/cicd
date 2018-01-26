all: push

TAG=${shell git describe --tags --long}
IMAGE_SERVER=index.boxlinker.com/boxlinker/boxci-server:${TAG}
IMAGE_AGENT=index.boxlinker.com/boxlinker/boxci-agent:${TAG}
push: push_server push_agent
push_server: build_server
	docker push ${IMAGE_SERVER}
push_agent: build_agent
	docker push ${IMAGE_AGENT}
build_server:
	cd cmd/server && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o boxci-server
	rm -rf release && mkdir -p release
	mv cmd/server/boxci-server release/
	cp cmd/server/.env.prod release/boxci-server.env
	docker build -t ${IMAGE_SERVER} .
build_agent:
	cd cmd/agent && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o boxci-agent
	rm -rf release && mkdir -p release
	mv cmd/agent/boxci-agent release/
	cp cmd/agent/.env.prod release/boxci-agent.env
	docker build -t ${IMAGE_AGENT} -f Dockerfile.agent .