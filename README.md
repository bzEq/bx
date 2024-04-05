# BX - ByteXchange
A simple and fast L4 virtual switch.

## Install
### Normal install
```shell
GOPROXY=direct go install github.com/bzEq/bx/i3@main
```
### Optimized install
```shell
GOPROXY=direct CXX=clang++ CGO_CXXFLAGS='-march=native -O3' \
    go install github.com/bzEq/bx/i3@main
```

## Kernel parameters
```conf
net.core.default_qdisc=cake
net.ipv4.tcp_congestion_control=bbr
```
