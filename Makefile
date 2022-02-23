APP_NAME=airquality
DOCKER_IMAGE=omarghader/airquality
SWARM_SERVICE=airquality_airquality

all: build 
arm64: build-arm64

compress:
	@upx -5 ./bin/* || true

build:
	@CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo -ldflags="-extldflags=-static" -o bin/${APP_NAME}

build-arm64:
	@CGO_ENABLED=0 GOARCH=arm64 go build -a -tags netgo -installsuffix netgo -ldflags="-extldflags=-static" -o bin/${APP_NAME}-arm64

docker-build:
	@docker buildx build --build-arg APP_NAME="${APP_NAME}" -t ${DOCKER_IMAGE}:latest .

docker-build-arm64:
	@docker buildx build --build-arg APP_NAME=${APP_NAME}-arm64 --platform linux/arm64 -t ${DOCKER_IMAGE}:arm64 .
