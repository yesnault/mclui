[![GoDoc](https://godoc.org/github.com/yesnault/mclui?status.svg)](https://godoc.org/github.com/yesnault/mclui)
[![Go Report Card](https://goreportcard.com/badge/yesnault/mclui)](https://goreportcard.com/report/yesnault/mclui)

# Marathon Command Line UI

Display applications deployed with marathon, with an UI based on https://github.com/gizak/termui.

![Screenshot](https://raw.githubusercontent.com/yesnault/mclui/master/screenshot.png)


# Usage

```
Usage:
  mclui [flags]
  mclui [command]

Available Commands:
  version     Display Version of mclui

Flags:
      --marathon-url stringSlice   URLs Marathon
      --with-auth-basic            Ask HTTP Basic Auth at startup (default true)

Use "mclui [command] --help" for more information about a command.
```

# Hacking

mclui is written in Go 1.7. Make sure you are using at least
version 1.7.

```bash
mkdir -p $GOPATH/src/github.com/yesnault
cd $GOPATH/src/github.com/yesnault
git clone git@github.com:yesnault/mclui.git
cd $GOPATH/src/github.com/yesnault/mclui
go build
```

You've developed a new cool feature? Fixed an annoying bug? We'd be happy
to hear from you! Make sure to read [CONTRIBUTING.md](./CONTRIBUTING.md) before.
