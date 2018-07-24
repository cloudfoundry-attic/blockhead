from golang:1.10-alpine3.8 as builder
workdir /go/src/github.com/jberkhahn/blockhead
copy . .
run go build -v -o /broker ./cmd/broker

from alpine:3.8
expose 3333
copy --from=builder /broker /broker
add ./creds.json /creds.json
cmd ["/broker","/creds.json"]