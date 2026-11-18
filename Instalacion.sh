#!/bin/bash

set -e

DIRECTORIO="$(cd "$(dirname "$0")" && pwd)"

cd "$DIRECTORIO"

echo "----------------------------------"
echo "    Instalación del proyecto      "
echo "----------------------------------"

for herramienta in go cmake; do
if ! command -v "$herramienta" >/dev/null 2>&1; then
    echo "Error: no se encontró $herramienta."
    echo "Instala los requisitos necesarios para continuar."
exit 1
fi
done

echo "Compilando el servidor..."

go build -o ROLAS-Servidor ./API

echo "Compilando la interfaz gráfica..."

cmake -S Interface/Interface -B Interface/Interface/build-release -DCMAKE_BUILD_TYPE=Release
cmake --build Interface/Interface/build-release --parallel

echo "Preparando el lanzador..."

chmod +x ROLAS.sh

echo ""
echo "---------------------------------------"
echo "  Instalación terminada correctamente  "
echo "---------------------------------------"
echo ""
echo "Para iniciar debes ejecutar:"
echo "./ROLAS.sh"
