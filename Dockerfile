FROM golang:1.22-alpine as builder
WORKDIR /
COPY . ./
RUN go mod download

RUN go build -o /grpc-mtl-server

FROM alpine
COPY --from=builder /grpc-mtl-server .

EXPOSE 7777
CMD [ "/grpc-mtl-server" ]