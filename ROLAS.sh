#1/bin/bash

DIRECTORIO_"$(cd "$(dirname "$0")" && pwd)"

cd "$DIRECTORIO"

if (echo > /dev/tcp/127.0.0.1/8080) 2>/dev/null; then
    SERVIDOR_PROPIO=0
else
    SERVIDOR_PROPIO=1

    ./ROLAS-Servidor &

    PID_SERVIDOR=$!

    trap 'kill "$PID_SERVIDOR" 2>/dev/null; wait "$PID_SERVIDOR" 2>/dev/null' EXIT

    for i in {1..50}; do
        if (echo > /dev/tcp/127.0.0.1/8080) 2>/dev/null; then
            break
        fi

        sleep 0.2
    done

    if ! (echo > /dev/tcp/127.0.0.1/8080) 2>/dev/null; then
        echo "No se pudo iniciar el servidor ROLAS."
        exit 1
    fi
fi

./Interface/Interface/build-release/Interface

