FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates tzdata
COPY . .
# 首次构建会自动拉取依赖并生成 go.sum
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/onehost .

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/onehost /app/onehost
COPY config/config.yaml /app/config/config.yaml
EXPOSE 8888
CMD ["/app/onehost"]
