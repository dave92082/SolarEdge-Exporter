FROM golang:1.26-alpine AS build
WORKDIR /see-build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o solaredge-exporter .

FROM alpine:3.22
ENV INVERTER_ADDRESS=192.168.1.189
ENV INVERTER_PORT=502
ENV EXPORTER_INTERVAL=5
RUN apk add --no-cache bash && adduser -D solaredge
# The home directory must stay writable: the default Log.Path is relative to
# the working directory.
WORKDIR /home/solaredge
COPY --from=build /see-build/solaredge-exporter .
USER solaredge
CMD ["./solaredge-exporter"]
