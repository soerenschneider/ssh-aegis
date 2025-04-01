FROM golang:1.24.2 AS build

ARG VERSION=dev
ARG COMMIT_HASH
ENV CGO_ENABLED=0

WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./

RUN go mod download
RUN CGO_ENABLED=${CGO_ENABLED} go build -ldflags="-w -X 'main.BuildVersion=${VERSION}' -X 'main.CommitHash=${COMMIT_HASH}'" -o /ssh-aegis .


FROM gcr.io/distroless/static AS final

LABEL maintainer="soerenschneider"
USER nonroot:nonroot
COPY --from=build --chown=nonroot:nonroot /ssh-aegis /ssh-aegis

ENTRYPOINT ["/ssh-aegis"]
