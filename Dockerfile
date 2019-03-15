# Start by building the application.
FROM golang:1.11 as build

WORKDIR /src/gitreleases

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN make install-go-deps
RUN make build

# Now copy it into our base image.
FROM gcr.io/distroless/base

COPY --from=build /src/gitreleases/gitreleases /
CMD ["/gitreleases"]
