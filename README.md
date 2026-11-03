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
- UPLOAD_DIR is optional and sets where uploaded files are stored (defaults to internal/uploads).
- MAX_UPLOAD is optional and sets the largest single upload request in GB (decimals allowed, defaults to 1). Files larger than this must be sent as fragments (see below).

Config loading is implemented in config.go.

## Usage

Start the server:

~~~bash
go run ./cmd/fs/main.go
~~~

Default upload directory (override with UPLOAD_DIR):
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
- 413 Request Entity Too Large if the file is bigger than MAX_UPLOAD. The size is checked before the body is read, so curl stops without sending the file.

### Fragmented upload (files larger than MAX_UPLOAD)

[cc-store](https://github.com/caml-cc/cc-store) does this automatically. The server appends fragments into a single file, so the finished upload is downloaded like any other.

1. Start the upload. The response has `Upload-Id` and `Upload-Max-Chunk` (max bytes per fragment) headers:

~~~bash
curl -i -X POST "http://localhost:5100/fragment" \
  -H "K: abc123" \
  -H "Upload-Length: 3000000000" \
  -H "Upload-Name: big.iso"
# optional: -H "Upload-Keep: 100h"   -H "Upload-Dir: 1" (marks a .tar.gz of a folder)
~~~

2. Send each fragment in order, as the raw request body, with the byte offset it starts at:

~~~bash
curl -i -X PATCH "http://localhost:5100/{id}" \
  -H "K: abc123" \
  -H "Upload-Offset: 0" \
  --data-binary @part0
~~~

Responses:
- 204 No Content with `Upload-Offset` (next offset) while the upload is incomplete
- 201 Created with the file URL once the last byte arrives
- 409 Conflict with `Upload-Offset` set to the server's actual offset if the offset is wrong (resume from there)
- 413 if a fragment is bigger than `Upload-Max-Chunk` or the remaining size

Incomplete uploads return 404 on download, show as `incomplete` in LIST, and expire like normal files.

### Download file

~~~bash
curl -L "http://localhost:5100/{id}" -o downloaded-file
~~~

Response:
- 200 OK
- binary file stream (supports Range requests)
- attachment filename comes from stored original filename
- `X-Folder: 1` header if the file is a folder uploaded by cc-store (a .tar.gz)

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
- 413 request entity too large when an upload exceeds MAX_UPLOAD
- 507 insufficient storage when an upload would exceed MAX_STORAGE
- 404 not found when file id does not exist
- 500 internal server error on storage/database failures

## Related projects

[cc-store](https://github.com/caml-cc/cc-store) CLI client for this API.