run:
	@go run ./cmd/fs/main.go

build:
	@go build -o ./bin/fs ./cmd/fs/main.go

clean:
	@rm ./bin/fs && rm -d ./bin/

full: build
	@./bin/fs
