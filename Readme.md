<img src="http://tjholowaychuk.com:6000/svg/title/NODE/PRUNE">

## What?

node-prune is a small tool to prune unnecessary files from ./node_modules, such as markdown, typescript source files, and so on. Primarily built for [Up](https://github.com/apex/up) which lets you deploy serverless web applications in seconds.

## Installation
If Go is already installed check check if your are using latest version, comparing to https://github.com/golang/go/releases, type this command to check your local Go version 

From source:

```
$ go version
```
Before install package make sure your GOPATH in` ~/.bash_profile` is set. It must be path to Go binaries. 

```
 export GOPATH="$HOME/go"
```

#### Install the package:
```
$ go get github.com/tj/node-prune/cmd/node-prune
```

From binary to `./bin/node-prune`:

```
$ curl -sfL https://install.goreleaser.com/github.com/tj/node-prune.sh | bash
```

From binary to `/usr/local/bin/node-prune`:

```
$ curl -sfL https://install.goreleaser.com/github.com/tj/node-prune.sh | bash -s -- -b /usr/local/bin
```

## Usage

In your app directory:

```
$ node-prune

files total 27,330
files removed 3,990
size removed 13 MB
   duration 200ms
```

Somewhere else:

```
$ node-prune path/to/node_modules

files total 27,330
files removed 3,990
size removed 13 MB
   duration 200ms
```

Or add to the ``package.json`` scripts field

```
  "scripts": {
    "postinstall": "node-prune"
  }
```

## Why?

![huge](https://pbs.twimg.com/media/DEIV_1XWsAAlY29.jpg)

---

[![GoDoc](https://godoc.org/github.com/tj/node-prune?status.svg)](https://godoc.org/github.com/tj/node-prune)
![](https://img.shields.io/badge/license-MIT-blue.svg)
![](https://img.shields.io/badge/status-stable-green.svg)

<a href="https://apex.sh"><img src="http://tjholowaychuk.com:6000/svg/sponsor"></a>
