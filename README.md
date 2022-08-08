# goversion

goversion is a simple version management tool.

## Concept

Since go maintains backward compatibility, it is possible to build past version sources with the latest go version.
Therefore, the concept of goversion is to make it easy to bring your local go version up to the latest version.

[There is also a way to install previous versions of go as a command](https://go.dev/doc/manage-install).
But this method is very forgettable. For this reason, goversion provides a command that makes it easy.

## Install

> **Note**
> Since go's install requires administrator rights,
> move it to a directory where administrators can run it.

```
go install github.com/komem3/goversion@latest && sudo mv $(go env GOPATH)/bin/goversion /usr/local/bin/
```

## Usage

### Upgrade go version

1. check latest version.

```
goversion -latest
```

2. upgrade version

```
sudo goversion -upgrade
```

### Install previous version

1. check version

```
goversion -ls-remote | grep 1.16
```

2. install specify version

```
goversion -install go1.16.15
```

## License

MIT
