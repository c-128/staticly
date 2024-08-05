FROM golang:1.22.4 AS build

WORKDIR /src
COPY . .
RUN go build -o /staticly server/*.go

FROM debian:latest AS run

ENV STATICLY_ADDRESS=0.0.0.0:3000
ENV STATICLY_ROOT="/data"
EXPOSE 3000
VOLUME [ "/data" ]

COPY --from=build /staticly /
CMD ["/staticly"]
