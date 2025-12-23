FROM golang:1.25.3-alpine3.22 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o main .

FROM alpine:3.14
# Create non-root user for OpenShift security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /home/appuser
COPY --from=builder /app/main .
RUN chown appuser:appgroup ./main

USER appuser
EXPOSE 8080
CMD ["./main"]
