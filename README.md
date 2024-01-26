# Object Storage Gateway
The gateway will store and retrieve objects on a set of minio instances running as docker containers.

- listening on port 3000/http by default
- **PUT** */object/{id}* - create new object with the specified ID and content body
- **GET** */object/{id}* - returns object content, or 400 if id doesn't exist.
- content types are not handled by the application
- *id* is the name of the object, and can be kind of whatever minio accepts.

Keep in mind that the *id* also defines which backing storage the object is stored to, having similar
*id*s stored to different backends.

## Quick start:
Assuming docker is installed, the following should:
- start the application stack
- create an object in one of the backing minio instances
- retrieve the from a backing minio instance
- show logs of the main application container and
- clean up (not completely...)

```bash
docker compose up -d
curl --location --request PUT 'localhost:3000/object/trala' --header 'Content-Type: text/plain' --data 'lala'
curl http://localhost:3000/object/trala
docker compose logs gateway-container
docker compose down
```
