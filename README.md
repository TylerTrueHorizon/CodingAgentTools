# Agent Tools Sandbox API

HTTP API that exposes agent-SDK-style tools (read, write, edit, shell, glob) as REST endpoints in a Dockerized Ubuntu service. Paths are not restricted to a workspace; shell runs commands as-is (including `sudo`).

## Build and run

### With Docker (recommended)

```bash
docker compose build
docker compose up
```

The API listens on port 8000.

**Security:** If `API_KEY` is set (e.g. in `.env` or your environment), every request must send the key via `Authorization: Bearer <key>` or `X-API-Key: <key>`. Generate a key with `openssl rand -hex 32` and set it in `.env` or pass it when running Docker (e.g. `API_KEY=your-key docker compose up`).

### Local (Go 1.22+)

```bash
go build -o sandbox-api ./cmd/server
./sandbox-api
```

Optional env: `PORT=8000`, `SHELL_TIMEOUT_SEC=120`, `MAX_REQUEST_BODY=10485760`, `API_KEY`. See `.env.example`.

---

## Endpoints

Base URL: `http://localhost:8000` (or your host/port).

### Files

**GET /files/read** — Read a file (optional line range).

```bash
curl -s "http://localhost:8000/files/read?path=/tmp/hello.txt"
curl -s "http://localhost:8000/files/read?path=/tmp/foo.txt&start_line=1&end_line=10"
```

**POST /files/write** — Create or overwrite a file.

```bash
curl -s -X POST http://localhost:8000/files/write \
  -H "Content-Type: application/json" \
  -d '{"path":"/tmp/hello.txt","content":"Hello, world.\n"}'
```

**POST /files/edit** — Edit a file (str_replace or insert at line).

```bash
# str_replace
curl -s -X POST http://localhost:8000/files/edit \
  -H "Content-Type: application/json" \
  -d '{"path":"/tmp/hello.txt","edit_type":"str_replace","old_str":"world","new_str":"sandbox"}'

# insert at line
curl -s -X POST http://localhost:8000/files/edit \
  -H "Content-Type: application/json" \
  -d '{"path":"/tmp/hello.txt","edit_type":"insert","line":1,"content":"# first line\n"}'
```

**GET /files/list** — List directory (optional glob pattern).

```bash
curl -s "http://localhost:8000/files/list?path=/tmp"
curl -s "http://localhost:8000/files/list?path=/tmp&pattern=*.txt"
# With API key (when API_KEY is set):
curl -s -H "Authorization: Bearer $API_KEY" "http://localhost:8000/files/list?path=/tmp"
```

### Shell

**POST /shell/run** — Run a shell command (e.g. `sh -c "..."`). Supports `sudo`.

```bash
curl -s -X POST http://localhost:8000/shell/run \
  -H "Content-Type: application/json" \
  -d '{"command":"whoami"}'

curl -s -X POST http://localhost:8000/shell/run \
  -H "Content-Type: application/json" \
  -d '{"command":"sudo apt-get update","timeout_seconds":120}'

curl -s -X POST http://localhost:8000/shell/run \
  -H "Content-Type: application/json" \
  -d '{"command":"ls -la","cwd":"/tmp"}'
```

---

## Docker details

- **Build**: Multi-stage; Go binary is built then copied into an Ubuntu 24.04 image.
- **User**: Container runs as root so that commands like `sudo apt-get install ...` work when sent via `POST /shell/run`.
- **Security**: Optional API key auth: set `API_KEY` at runtime; clients send it via `Authorization: Bearer <key>` or `X-API-Key: <key>`.
