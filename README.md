# Kappy

Kappy is an opinionated Kubernetes build and deployment helper.

If you have one or more services that make up a part of your stack, or even your entire application, you can use Kappy to orchestrate CI, building and deployment tasks for one or more individual services.

# Configuration

### `docker_registry` 

You can define an alternative `docker_registry`.  This will then be included in an image's tags where the image is defined in the `services` section along with a `build` section, if `build` section is not sepcified you must fully qualify your image names if using a private registry.  

For example, if you have `docker_registry: dkr.myregistry.com`, and an `image` with a `build` section in the `services.<name>.containers` section, where `image: my/image`, a tag will be added to the built image like: `dkr.myregistry.com/my/image` so that it can be pushed using `docker push` to that registry. 

In `.kappy.yaml`:

```YAML
docker_registry: dkr.example.com
```

### `services`

You can define a set of `services` to be managed by `kappy` in this section. `services` is a map of services that will correspond to the service's unique name in Kubernetes.

In `.kappy.yaml`:

```YAML
services:
  frontend-nginx:
```

#### `containers`

An instance of a service is defined by a set of containers that make up a pod in kubernetes.  A service can contain any number of containers which will be scheduled together.  In future versions of Kappy you'll be able to define resource shares so that pods can share the pid namespace, or a shared volume. 

Each container is a running instance of a docker image.  Inside the container section you can define which image to use. Additionally you can specify how Kappy can build this for you.

```YAML
services:
  service-name:
    containers:
      - name: container-name
        build:
          # ...
        image: my/image-name
```

If a `build` section is specified, the image will be tagged with the `image` name defined, and if the `docker_registry` key is defined additionally, this will be used to tag the image.

##### `build`

This section is inspired by `docker-compose`'s `build` section: https://docs.docker.com/compose/compose-file/#build

```YAML
services:
  service-name:
    containers:
      - name: container-name
        build:
          context: .
          dockerfile: api/myapi/Dockerfile
          args:
            buildNo: 1
```

// TODO
