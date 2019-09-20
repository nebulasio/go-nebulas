module github.com/nebulasio/go-nebulas

go 1.12

require (
	github.com/btcsuite/btcd v0.0.0-20190523000118-16327141da8c
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/gogo/protobuf v1.3.0
	github.com/golang/protobuf v1.3.2
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/grpc-gateway v1.11.2
	github.com/hashicorp/golang-lru v0.5.3
	github.com/influxdata/influxdb v1.7.8
	github.com/lestrrat-go/file-rotatelogs v2.2.0+incompatible
	github.com/lestrrat-go/strftime v0.0.0-20190725011945-5c849dd2c51d // indirect
	github.com/libp2p/go-libp2p v0.3.1
	github.com/libp2p/go-libp2p-crypto v0.1.0
	github.com/libp2p/go-libp2p-host v0.1.0
	github.com/libp2p/go-libp2p-kbucket v0.2.1
	github.com/libp2p/go-libp2p-net v0.1.0
	github.com/libp2p/go-libp2p-peer v0.2.0
	github.com/libp2p/go-libp2p-peerstore v0.1.3
	github.com/libp2p/go-libp2p-swarm v0.2.1
	github.com/multiformats/go-multiaddr v0.0.4
	github.com/peterh/liner v1.1.0
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/robertkrimen/otto v0.0.0-20180617131154-15f95af6e78d
	github.com/rs/cors v1.7.0
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/syndtr/goleveldb v1.0.0
	github.com/tecbot/gorocksdb v0.0.0-20190705090504-162552197222
	github.com/urfave/cli v1.22.1
	github.com/willf/bitset v1.1.10 // indirect
	github.com/willf/bloom v2.0.3+incompatible
	golang.org/x/crypto v0.0.0-20190618222545-ea8f1a30c443
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc v1.22.4-0.20190911181634-40c00c22f2e7
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
)

replace github.com/grpc-ecosystem/grpc-gateway => github.com/nebulasio/grpc-gateway v1.11.2
