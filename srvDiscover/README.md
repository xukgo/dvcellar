# etcd依赖问题


go get google.golang.org/grpc@v1.26.0

google.golang.org/grpc v1.26.0   (改成v1.26.0版本)

在go.mod中添加如下两行

replace github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4

replace go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4




