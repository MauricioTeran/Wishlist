# Microservicio de Wishlist

Microservicio REST desarrollado en Go encargado de gestionar la lista de vuelos de interés (wishlist) de los usuarios dentro de la plataforma de viajes.

Este servicio permite almacenar de manera temporal los vuelos seleccionados por cada usuario, sin utilizar un sistema de persistencia tradicional.

## Tecnologías

* Go
* net/http (librería estándar)

## Descripción

El microservicio de wishlist permite a los usuarios guardar vuelos de interés para consultas futuras.

A diferencia de otros microservicios del sistema, este componente no utiliza una base de datos, sino que almacena la información en memoria utilizando estructuras concurrentes.

La información se organiza como una relación entre el identificador del usuario y los vuelos asociados a su lista.

## Modelo de datos

La estructura principal utilizada es un mapa en memoria:

* Usuario → Lista de vuelos

Ejemplo conceptual:

```json id="wishlist_example"
{
  "usuarioId": 1,
  "vueloId": 10
}
```

## Características

* Almacenamiento en memoria (in-memory)
* Uso de estructuras concurrentes para acceso seguro
* Operaciones rápidas de lectura y escritura
* No persistente

## Endpoints

### Agregar elemento

* POST /wishlist

Permite agregar un vuelo a la wishlist de un usuario.

### Obtener wishlist

* GET /wishlist/{userId}

Devuelve la lista de vuelos asociados a un usuario.
Este endpoint consulta el microservicio de tickets para obtener los datos completos de cada vuelo.

### Eliminar elemento

* DELETE /wishlist

Elimina un vuelo específico de la wishlist.

### Vaciar wishlist

* DELETE /wishlist/{userId}/clear

Elimina todos los elementos de la wishlist de un usuario.

### Health Check

* GET /health

Devuelve el estado del servicio.

## Variables de entorno

```env id="wishlist_env"
TICKETS_SERVICE_URL=http://localhost:8000
```

## Ejecución del proyecto

### 1. Ejecutar el servicio

```bash id="wishlist_run"
go run main.go
```

El servicio se ejecuta por defecto en el puerto 8082.

## Integración con otros microservicios

Este microservicio depende directamente del microservicio de tickets:

* Consulta vuelos mediante: GET /vuelos/{id}

Esto permite obtener información actualizada sin duplicar datos.

## Flujo de funcionamiento

1. El usuario agrega un vuelo a su wishlist.
2. El sistema almacena el identificador del vuelo en memoria.
3. Cuando se consulta la wishlist, el servicio solicita los datos al microservicio de tickets.
4. Se devuelve la información completa al cliente.

## Limitaciones

* No existe persistencia de datos
* La información se pierde al reiniciar el servicio
* No está diseñado para almacenamiento a largo plazo
* Escalabilidad limitada sin mecanismos adicionales

## Justificación de diseño

El uso de almacenamiento en memoria permite:

* Reducir la latencia
* Simplificar la implementación
* Evitar dependencia de bases de datos
* Mantener una arquitectura desacoplada

Este enfoque es adecuado para datos temporales y no críticos.

## Rol dentro del sistema

Este microservicio permite mejorar la experiencia del usuario al ofrecer una funcionalidad de selección de vuelos de interés, integrándose con el microservicio de tickets para obtener información en tiempo real.

## Notas

* No se duplica información de vuelos
* Se mantiene una única fuente de verdad en el microservicio de tickets
* Diseñado para integrarse en una arquitectura distribuida basada en microservicios
