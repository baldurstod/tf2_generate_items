.PHONY: build clean

BINARY_NAME=tf2_generate_items

build:
	go build ${BINARY_NAME}

run: build
	./${BINARY_NAME} -o "./var" -i "./var" -r "./var" -m=0 -l english

clean:
	go clean
