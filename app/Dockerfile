FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG GIT_SHA=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${GIT_SHA}" \
    -o /out/items-api ./cmd/items-api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/items-api /items-api
EXPOSE 8000
USER 65532:65532
ENTRYPOINT ["/items-api"]
