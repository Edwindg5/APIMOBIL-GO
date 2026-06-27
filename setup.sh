#!/bin/bash
# Script de configuración inicial para api-mobile
# Este script prepara el entorno para ejecutar la aplicación

set -e

echo "=========================================="
echo "Kajve API Mobile - Setup Inicial"
echo "=========================================="
echo ""

# Verificar si Go está instalado
if ! command -v go &> /dev/null; then
    echo "ERROR: Go no está instalado. Por favor instala Go 1.23 o superior"
    exit 1
fi

echo "✓ Go instalado: $(go version)"
echo ""

# Verificar si Docker está instalado
if ! command -v docker &> /dev/null; then
    echo "WARNING: Docker no está instalado. Se necesita para correr PostgreSQL y Redis"
    echo "Puedes descargar Docker desde: https://www.docker.com/products/docker-desktop"
fi

# Crear archivo .env si no existe
if [ ! -f .env ]; then
    echo "Creando archivo .env..."
    cp .env.example .env
    echo "✓ Archivo .env creado (revisa y actualiza las variables sensibles)"
else
    echo "✓ Archivo .env ya existe"
fi

echo ""
echo "Descargando dependencias..."
go mod download
go mod tidy
echo "✓ Dependencias descargadas"

echo ""
echo "=========================================="
echo "Próximos pasos:"
echo "=========================================="
echo ""
echo "1. Para ejecutar con Docker (recomendado):"
echo "   docker-compose up -d"
echo "   go run ./cmd/main.go"
echo ""
echo "2. Para ejecutar localmente (requiere PostgreSQL y Redis):"
echo "   - Asegúrate de que PostgreSQL 16+ esté corriendo en localhost:5432"
echo "   - Asegúrate de que Redis esté corriendo en localhost:6379"
echo "   - go run ./cmd/main.go"
echo ""
echo "3. Para compilar:"
echo "   go build -o bin/api-mobile ./cmd/main.go"
echo ""
echo "4. Para ver todos los comandos disponibles:"
echo "   make help"
echo ""
echo "5. Base de datos inicial:"
echo "   - Las tablas se crean automáticamente con init.sql en Docker"
echo "   - Necesitas ejecutar scripts/init.sql manualmente si usas BD local"
echo ""
echo "=========================================="
echo "Setup completado ✓"
echo "=========================================="
