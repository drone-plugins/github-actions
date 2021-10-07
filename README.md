# drone-github-action-plugin

This plugin allows running github actions as a drone plugin.

## Build

Build the binaries with the following commands:

```console
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
export GO111MODULE=on

go build -v -a -tags netgo -o release/linux/amd64/plugin   ./cmd

```

## Docker

Build the Docker images with the following commands:

```console
docker build \
  --label org.label-schema.build-date=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --label org.label-schema.vcs-ref=$(git rev-parse --short HEAD) \
  --file docker/Dockerfile.linux.amd64 --tag plugins/github-actions .

```

## Plugin step usage

Provide uses, with & env of github action to use in plugin step settings. Provide GITHUB_TOKEN as environment variable if it is required for an action.

```console
steps:
- name: github-action
  image: plugins/github-actions
  settings:
    uses: actions/hello-world-javascript-action@v1.1
    with:
        who-to-greet: Mona the Octocat
    env:
        hello: world

```

## Running locally

1. Running actions/hello-world-javascript-action action locally via docker:

```console

 docker run --rm \
    --privileged \
    -v $(pwd):/drone \
    -w /drone \
    -e PLUGIN_USES="actions/hello-world-javascript-action@v1.1" \
    -e PLUGIN_WITH="{\"who-to-greet\":\"Mona the Octocat\"}" \
    -e PLUGIN_VERBOSE=true \
    plugins/github-actions

```
