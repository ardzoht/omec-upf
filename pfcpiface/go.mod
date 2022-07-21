module github.com/ardzoht/omec-upf/pfcpiface

go 1.16

require (
	github.com/Showmax/go-fqdn v1.0.0
	github.com/golang/protobuf v1.5.2
	github.com/google/gopacket v1.1.19
	github.com/libp2p/go-reuseport v0.2.0
	github.com/omec-project/pfcpsim v0.0.0-20220328122841-64474e93876e
	github.com/omec-project/upf-epc v0.3.0
	github.com/prometheus/client_golang v1.12.2
	github.com/stretchr/testify v1.7.2
	github.com/wmnsk/go-pfcp v0.0.15
	go.uber.org/zap v1.21.0
	google.golang.org/genproto v0.0.0-20220608133413-ed9918b62aac // indirect
	google.golang.org/grpc v1.47.0
	google.golang.org/protobuf v1.28.0
)

retract [v0.3.0, v0.6.9] // These versions are deprecated
