BINARY_NAME=build/ssh_cert_ca
ARCH=arm64

.PHONY: clean build
build:
	GOOS=linux GOARCH=${ARCH} CGO_ENABLE=0 go build -v -ldflags=" -s -w" -o ${BINARY_NAME} main.go
	# upx -9 -f ${BINARY_NAME}
run:
	./${BINARY_NAME}


publish:
	GOOS=linux CGO_ENABLE=0 GOARCH=${ARCH} go build -ldflags=" -s -w" -o ${BINARY_NAME} main.go
	upx -9 -f ${BINARY_NAME}

clean:
	go clean || rm -f ${BINARY_NAME}
