# fs

Small self-hosted file API for uploading, downloading, and deleting files.

## Install

### Option 1: Run from source (recommended)

~~~bash
go run ./cmd/fs/main.go
~~~

### Option 2: Build binary

~~~bash
go build -o ./bin/fs ./cmd/fs/main.go
./bin/fs
~~~

### Option 3: Using make targets

~~~bash
make run
make build
make full
~~~

See makefile for available commands.

## Config

Create a .env file in the project root:

~~~env
PORT=5100
API_KEY=abc123
~~~

You can copy from [ .env.example ](.env.example) and update values.

Notes:
- PORT is required.
- API_KEY is required and used in request header K for protected routes.

Config loading is implemented in config.go.

## Usage

Start the server:

~~~bash
go run ./cmd/fs/main.go
~~~

Default upload directory:
- internal/uploads

Database:
- SQLite at fs.db
- migrations in internal/database/migrations
- DB init in sqlite.go

## API

Routes are registered in server.go.

### Upload file

~~~bash
curl -X POST "http://localhost:5100/" \
  -H "K: abc123" \
  -F "file=@./example.txt"
~~~

Response:
- 201 Created
- plain text URL with generated file id, for example:
  localhost:5100/AbC123xYz9

### Download file

~~~bash
curl -L "http://localhost:5100/{id}" -o downloaded-file
~~~

Response:
- 200 OK
- binary file stream
- attachment filename comes from stored original filename

### Delete file

~~~bash
curl -X DELETE "http://localhost:5100/{id}" \
  -H "K: abc123"
~~~

Response:
- 204 No Content

## Errors

Typical responses:
- 401 unauthorized when K header is missing or invalid
- 400 bad request when required input is missing
- 404 not found when file id does not exist
- 500 internal server error on storage/database failures

## Related projects

[cc-store](https://github.com/caml-cc/cc-store) CLI client for this API.