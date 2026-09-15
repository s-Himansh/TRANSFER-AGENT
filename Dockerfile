FROM node:20-alpine AS ui-builder
WORKDIR /app/ui
COPY ui/package*.json ./
RUN npm ci
COPY ui/ .
RUN npm run build

FROM golang:1.22-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui-builder /app/ui/.next/static ./ui/.next/static
COPY --from=ui-builder /app/ui/public ./ui/public
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/transfer-agent .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=go-builder /app/transfer-agent .
RUN mkdir -p generated_file
EXPOSE 8080 6789
ENTRYPOINT ["./transfer-agent"]
CMD ["-mode", "web"]
