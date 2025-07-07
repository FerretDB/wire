FROM golang:1.24.4

RUN <<EOF
go install -v golang.org/x/tools/gopls@latest
go install -v github.com/go-delve/delve/cmd/dlv@latest

go clean -cache
EOF
