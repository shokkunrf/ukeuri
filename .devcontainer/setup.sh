#!/bin/bash

go get -v \
  github.com/uudashr/gopkgs/v2/cmd/gopkgs \
  github.com/ramya-rao-a/go-outline \
  github.com/cweill/gotests/gotests \
  github.com/fatih/gomodifytags \
  github.com/josharian/impl \
  github.com/haya14busa/goplay/cmd/goplay \
  github.com/go-delve/delve/cmd/dlv \
  honnef.co/go/tools/cmd/staticcheck \
  golang.org/x/tools/gopls

go mod tidy
