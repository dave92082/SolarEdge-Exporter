FROM golang:1.12 AS build
RUN mkdir /see-build
ADD . /see-build/
WORKDIR /see-build
RUN CGO_ENABLED=0 go build -o solaredge-exporter .

FROM alpine:latest
ENV INVERTER_ADDRESS 192.168.1.189
ENV INVERTER_PORT 502
ENV EXPORTER_INTERVAL 5
RUN apk add --no-cache bash
WORKDIR /root
COPY --from=build /see-build/solaredge-exporter .
CMD ["./solaredge-exporter"]