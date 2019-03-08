# Start by building the application.
FROM golang:1.11 as build

WORKDIR /go/src/gitreleases
COPY . .

RUN go get -d -v ./...
RUN make build

# Now copy it into our base image.
FROM gcr.io/distroless/base

COPY --from=build /go/src/gitreleases/gitreleases /
CMD ["/gitreleases"]
