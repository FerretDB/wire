FROM golang:1.24.5

COPY go.mod go.sum /tmp/wire/
COPY tools/go.mod tools/go.sum /tmp/wire/tools/

RUN <<EOF
go install -v golang.org/x/tools/gopls@latest
go install -v github.com/go-delve/delve/cmd/dlv@latest
go clean -cache

cd /tmp/wire
go mod download -x
go -C tools mod download -x
rm -fr /tmp/wire
EOF
