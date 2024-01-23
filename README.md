# Object Storage Gateway
The gateway will store and retrieve objects on a set of minio instances running as docker containers.

- listening on port 3000/http by default
- **PUT** */object/{id}* - create new object with the specified ID and content body
- **GET** */object/{id}* - returns object content, or 404 if id doesn't exist.
- *id* is alphanumeric, up to 32 characters.
