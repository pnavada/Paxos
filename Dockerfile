FROM golang:1.21-alpine

ENV GOPATH=/go
ENV PATH=$PATH:/go/bin

WORKDIR /go/src/paxos

COPY . .

RUN go mod tidy

ENTRYPOINT [ "go", "run", "main.go" ]