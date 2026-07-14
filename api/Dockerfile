FROM golang:1.24-alpine AS build

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/blog-api ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache tzdata \
	&& cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
	&& echo "Asia/Shanghai" > /etc/timezone
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=build /bin/blog-api /usr/local/bin/blog-api
COPY --from=build /app/email-templates ./email-templates
EXPOSE 8080
CMD ["blog-api"]
