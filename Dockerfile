FROM docker.io/library/golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
COPY pkg ./pkg
RUN mkdir -p /out/app && mkdir -m 1777 /out/data \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app/scheduler .

FROM scratch
WORKDIR /app
COPY --from=build /out/ /
COPY web /app/web
USER 65532:65532
ENV TODO_PORT=7540 TODO_DBFILE=/data/scheduler.db
EXPOSE 7540
ENTRYPOINT ["/app/scheduler"]
