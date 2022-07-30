/*

Command goversion updates your go version.

	$ go install github.com/komem3/goversion@latest

First, check latest go version.

	goversion -latest

If your go version is older than latest version, you can upgrade with the following command.

	sudo $(go env GOPATH)/bin/goversion -upgrade

*/
package main
