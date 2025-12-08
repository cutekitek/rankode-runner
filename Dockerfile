FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download 2>/dev/null || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/worker cmd/main.go

FROM alpine:latest

ARG LANGUAGES="c,c++,go,java21,python,javascript,c#,perl,rust"
ENV LANGUAGES=${LANGUAGES}

ENV RUSTUP_HOME=/usr/local/rustup \
    CARGO_HOME=/usr/local/cargo \
    PATH="/usr/local/cargo/bin:/usr/lib/go/bin:$PATH" \
    DOTNET_SYSTEM_GLOBALIZATION_INVARIANT=1

COPY --from=builder /app/worker /app/worker
COPY install_compilers.sh /app/install_compilers.sh
COPY languages /app/languages

RUN chmod +x /app/install_compilers.sh && \
    /app/install_compilers.sh

WORKDIR /app
CMD ["./worker"]