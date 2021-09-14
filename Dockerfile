FROM golang:1.17 AS stage1-base
WORKDIR /go/cmd/app

FROM stage1-base AS stage2-deps
COPY go.mod .
COPY go.sum .
COPY Makefile .
RUN make install

FROM stage2-deps As stage3-build
COPY cmd/wptest cmd/wptest
RUN make build

FROM gcr.io/distroless/base:latest AS stage4-distroless 
COPY --from=stage3-build /go/cmd/app/bin/app .
COPY configs configs
ENTRYPOINT ["./app"]