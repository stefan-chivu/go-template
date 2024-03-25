This repo contains a template for a basic Golang REST API server supporting Websocket connections.

After cloning, use:
```
find . -type f -not -path '*/\.*' -exec sed -i 's/go-template/<package-name>/g' {} +
```
And then:
```
mv go-template <package-name>
```
