FROM golang:1.21.6 as builder

WORKDIR /workspace

COPY . . 

RUN go env -w GOPROXY=https://goproxy.cn

RUN go mod tidy

RUN CGO_ENABLED=0 go build -ldflags='-s -w' -o ingress-validator main.go

FROM alpine:3.19.1

COPY --from=builder /workspace/ingress-validator /ingress-validator

ENTRYPOINT ["/ingress-validator"]