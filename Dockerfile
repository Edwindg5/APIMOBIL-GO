FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copiar archivos del módulo
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-mobile .

# Imagen final
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiar binario desde el builder
COPY --from=builder /app/api-mobile .
COPY --from=builder /app/.env.example .

EXPOSE 8080

CMD ["./api-mobile"]
