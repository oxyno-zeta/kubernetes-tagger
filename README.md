# kubernetes-tagger

[![CircleCI](https://circleci.com/gh/oxyno-zeta/kubernetes-tagger/tree/master.svg?style=svg)](https://circleci.com/gh/oxyno-zeta/kubernetes-tagger/tree/master) [![Go Report Card](https://goreportcard.com/badge/github.com/oxyno-zeta/kubernetes-tagger)](https://goreportcard.com/report/github.com/oxyno-zeta/kubernetes-tagger) ![Docker Pulls](https://img.shields.io/docker/pulls/oxynozeta/kubernetes-tagger.svg)

## Context

Kubernetes tagger offer the possibility to add tags on external services like EBS on AWS.

Why creating this project ? Because Kubernetes doesn't offer this feature for the moment.

## How it works ?

For persistent volume, Kubernetes-tagger will watch for new persistent volumes and test if they can be processed.

When they can be processed, it will test if rules can be applied with actual tags present and Kubernetes data.

Once, this is done, kubernetes-tagger will apply the delta on the target provider.

## Documentation

- [Configuration](docs/configuration.md)
- [Data Structure](docs/data-structure.md)

## Contributing

Of course, we accept contribution on the project.

To contribute on it, just read the [contributing guidelines](CONTRIBUTING.md).
