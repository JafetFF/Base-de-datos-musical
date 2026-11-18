Modelado y Programación
=======================

Proyecto 2: Base de datos musical
-----------------------

### Fecha de entrega: viernes 20 de noviembre, 2026

Este proyecto tiene 4 componentes principales: una base
de datos; un mini-compilador; un minero; y una interface.
La base de datos de música es almacenada con ayuda del
minero que extrae los archivos mp3, así también como sus metadatos de manera recursiva en un directorio.

### Acerca del proyecto ROLAS

ROLAS es una base de datos musical para administrar tus propias rolas,
donde se recopila información de archivos de audio MP3, se organizan
rolas, álbumes y artistas, se realizan búsquedas y también se reproduce
la música desde una interface gráfica.

-----------------------

### Características de cada componente

- Minero: recorre directorios recursivamente y extrae los metadatos de los archivos MP3.
- BDD: almacena y organiza rolas, álbumes y artistas.
- Gestión de artistas: permite añadir personas y grupos, editar su información y administrar integrantes de los grupos.
- Gestión de álbumes: se permite añadir, consultar y editar álbumes.
- Edición de rolas: es posible modificar la información almacenada de las rolas en la BDD.
- Búsquedas: se puede consultar la biblioteca mediante el mini-compiladorcito.
- Reproductor: se pueden reproducir los archivos de audio desde la aplicación.
- Portadas: muestra las portadas de los álbumes cuando existen portadas.

### Tecnologías utilizadas

- Go (Para la mayor parte del proyecto: el compilador; la API; el Modelo; el Servidor; y el Minero)
- SQLite (Para la BDD)
- C++ (Para la parte lógica de la interface)
- Qt 6 (Para la interface)

### Requisitos para compilar y ejecutar ROLAS

- Go (versión compatible con la que está en "go.mod")
- CMake 3.16 
- Compilador de C++ 
- Qt 6

### Instalación y ejecución

1. Clonamos el repositorio

```
$ git clone https://github.com/JafetFF/Base-de-datos-musical.git
```

Y entramos al directorio del proyecto:

```
$ cd Base-de-dato-musical
```

2. Instalar ROLAS

Damos permisos de ejecución al instalador:

```
$ chmod +x instalar.sh
```

Ejecutamos:

```
$ ./instalar.sh
```

El instalador se encarga de compilar automáticamente el servidor y la interface gráfica

3. Ejecutar ROLAS

Después de que finalice la instalación, abrimos ROLAS:

```
$ ./ROLAS.sh
```

### Respecto a la BDD

La primera vez que se ejecuta ROLAS, el servidor crea automáticamente la BDD SQLite si todavía no existe, utilizando el esquema del proyecto.

La BDD se almacena en:

```
BDD/BaseMusical.db
```

### Estructura del proyecto

Directorio / archivo| Descripción
"API/"| Servidor y API para la comunicación con la interface.
"BDD/"| Esquema SQL y almacenamiento de la base de datos y de las portadas.
"Compilador/"| Implementación del mini-compiladorcito.
"Interface/"| Interface gráfica.
"Minero/"| Donde está el la implementación del minero.
"Modelo/"| Modelo de los datos y operaciones sobre la BDD.
"ROLAS.sh"| Lanzador de la BDD musical.

