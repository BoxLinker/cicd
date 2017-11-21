
PREFIX=index.boxlinker.com/boxlinker

IMAGE_EMAIL=email-server
IMAGE_EMAIL_TAG=latest

IMAGE_ID=user-server
IMAGE_ID_TAG=latest

IMAGE_PREFIX=index.boxlinker.com/boxlinker
IMAGE_ALIYUN_PREFIX=registry.cn-beijing.aliyuncs.com/cabernety
IMAGE_REGISTRY=registry-server
IMAGE_REGISTRY_TAG=v1.0

IMAGE_USER=user-server
IMAGE_USER_TAG=v1.0

IMAGE_APP=application-server
IMAGE_APP_TAG=${shell git describe --tags --long}

IMAGE_ROLL=rolling-update
IMAGE_ROLL_TAG=v1.0

IMAGE_REG_WAT=registry-watcher
IMAGE_REG_WAT_TAG=v1.0


db:
	docker rm -f boxlinker-db-test || true
	docker run -d --name boxlinker-db-test -v `pwd`/db_data:/var/lib/mysql -p 3306:3306 -e MYSQL_DATABASE=boxlinker -e MYSQL_ROOT_PASSWORD=123456 mysql

rabbitmq:
	docker rm -f boxlinker-email-rabbitmq || true
	docker run -d --name boxlinker-email-rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

build-registry:
	cd cmd/registry && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o image-api
	docker build -t ${IMAGE_ALIYUN_PREFIX}/${IMAGE_REGISTRY}:${IMAGE_REGISTRY_TAG} -f Dockerfile.registry .

registry: build-registry
	docker push ${IMAGE_ALIYUN_PREFIX}/${IMAGE_REGISTRY}:${IMAGE_REGISTRY_TAG}


email: push-email

build-email:
	cd cmd/email && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o email
	docker build -t ${IMAGE_ALIYUN_PREFIX}/${IMAGE_EMAIL}:${IMAGE_EMAIL_TAG} -f Dockerfile.email .

push-email: build-email
	docker push ${IMAGE_ALIYUN_PREFIX}/${IMAGE_EMAIL}:${IMAGE_EMAIL_TAG}

build-user:
	cd cmd/user && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o user
	docker build -t ${IMAGE_ALIYUN_PREFIX}/${IMAGE_ID}:${IMAGE_USER_TAG} -f Dockerfile.user .

push-user: build-user
	docker push ${IMAGE_ALIYUN_PREFIX}/${IMAGE_ID}:${IMAGE_USER_TAG}

user: push-user

build-application:
	cd cmd/application && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o application
	docker build -t ${IMAGE_PREFIX}/${IMAGE_APP}:${IMAGE_APP_TAG} -f Dockerfile.application .

application: build-application
	docker push ${IMAGE_PREFIX}/${IMAGE_APP}:${IMAGE_APP_TAG}

build-rolling-update:
	cd cmd/rolling-update && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o rolling-update
	docker build -t ${IMAGE_PREFIX}/${IMAGE_ROLL}:${IMAGE_ROLL_TAG} -f Dockerfile.rolling-update .

rolling-update: build-rolling-update
	docker push ${IMAGE_PREFIX}/${IMAGE_ROLL}:${IMAGE_ROLL_TAG}

build-registry-watcher:
	cd cmd/registry-watcher && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o registry-watcher
	docker build -t ${IMAGE_PREFIX}/${IMAGE_REG_WAT}:${IMAGE_REG_WAT_TAG} -f Dockerfile.registry-watcher .

registry-watcher: build-registry-watcher
	docker push ${IMAGE_PREFIX}/${IMAGE_REG_WAT}:${IMAGE_REG_WAT_TAG}

minikube:
	minikube start --kubernetes-version=v1.6.0 --extra-config=kubelet.PodInfraContainerImage="registry.cn-beijing.aliyuncs.com/cabernety/pause-amd64:3.0" --registry-mirror="2h3po24q.mirror.aliyuncs.com"