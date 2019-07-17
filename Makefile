GOC=go
BUILT_DIR=./bin

cleanup:
	@rm -rf ${BUILT_DIR}

build: cleanup
	@${GOC} build -o ${BUILT_DIR}/main ./main.go

run: build
	@${BUILT_DIR}/main
