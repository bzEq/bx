# BX - ByteXchange
A simple and fast L4 virtual switch.

## Install
### Normal install
```shell
GOPROXY=direct go install github.com/bzEq/bx/i3@main
```
### Optimized install
```shell
GOPROXY=direct CXX=clang++ CGO_CXXFLAGS='-march=native' CGO_LDFLAGS='-lmimalloc' \
    go install github.com/bzEq/bx/i3@main
```
