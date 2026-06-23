# Rankode Runner

`rankode-runner` is the worker service that executes submitted code for Rankode. It listens to RabbitMQ queue `rankode-req`, downloads test/verification files from S3-compatible storage, runs code in sandboxed language environments, and publishes results to `rankode-resp`.

## How It Works

1. The backend receives a code attempt.
2. The backend sends an execution request to RabbitMQ.
3. Runner workers consume requests from `rankode-req`.
4. The runner loads task files from S3-compatible storage.
5. Code is executed with the language config from `languages/<language>/config.json`.
6. The result is sent back through `rankode-resp`.

## Supported Languages

The Docker image can install only the compilers/interpreters you need. Supported values for `LANGUAGES` include:

- `python3` or `python`
- `go` or `golang`
- `c`
- `c++` or `cpp`
- `java` or `java21`
- `javascript`, `js`, or `bun`
- `rust` or `rustlang`
- `csharp`, `c#`, or `dotnet`
- `perl`

Language runtime scripts and limits live in the `languages/` directory.

## Configuration

The service reads `.env` if it exists, otherwise environment variables.

Example `.env`:

```env
RABBIT_HOST=rabbitmq
RABBIT_PORT=5672
RABBIT_USER=rankode
RABBIT_PASSWORD=<password>

S3_ENDPOINT=seaweedfs-s3:8333
S3_ACCESS_KEY=rankode
S3_SECRET_KEY=<password>
S3_BUCKET=tasks

WORKERS_COUNT=0
LOG_LEVEL=info
LANGUAGES=python3,go
```

`WORKERS_COUNT=0` means the runner uses the number of CPU cores.

## Build Docker Image With Required Languages

Build an image with Python and Go support:

```bash
docker build \
  --build-arg LANGUAGES=python3,go \
  -t rankode-runner:python-go .
```

Build an image with a wider toolchain:

```bash
docker build \
  --build-arg LANGUAGES=python3,go,java,c,c++,javascript \
  -t rankode-runner:full .
```

Rust, JavaScript/Bun, and some SDKs are downloaded during image build, so network access is required for those language sets.

## Run With Docker Compose

The compose file runs the worker with privileges required by the sandbox and mounts cgroups:

```bash
docker compose up -d --build
```

To choose languages through compose, set `LANGUAGES` in `.env` before building:

```env
LANGUAGES=python3,go,java
```

Then rebuild:

```bash
docker compose build --no-cache
docker compose up -d
```

View logs:

```bash
docker compose logs -f rankode-runner
```

Stop the worker:

```bash
docker compose down
```

## Run With `docker run`

The runner must be on the same Docker network as RabbitMQ and SeaweedFS from the backend stack. Replace `rankode_rankode` with the actual network name if it differs.

```bash
docker run --rm \
  --name rankode-runner \
  --privileged \
  --cgroupns=host \
  --network rankode_rankode \
  --tmpfs /tmp:exec,mode=1777,size=1g \
  -v /sys/fs/cgroup:/sys/fs/cgroup:rw \
  -e RABBIT_HOST=rabbitmq \
  -e RABBIT_PORT=5672 \
  -e RABBIT_USER=rankode \
  -e RABBIT_PASSWORD=<password> \
  -e S3_ENDPOINT=seaweedfs-s3:8333 \
  -e S3_ACCESS_KEY=rankode \
  -e S3_SECRET_KEY=<password> \
  rankode-runner:python-go
```

## Local Development

Run unit tests:

```bash
go test ./...
```

Sandbox tests and benchmarks require elevated permissions:

```bash
sudo go test -v ./internal/runner/sandbox
sudo go test -bench=. ./benchmarks
```
