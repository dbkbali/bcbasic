build:
	@go build -o bin/bcbasic  -v

run: build
	@./bin/bcbasic

test:
	@go test -v ./... -count=1
