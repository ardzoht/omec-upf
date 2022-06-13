module github.com/ardzoht/omec-upf

go 1.16

require (
	github.com/ettle/strcase v0.1.1
	github.com/golang/protobuf v1.5.2
	github.com/omec-project/upf-epc v0.3.0
	github.com/p4lang/p4runtime v1.3.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.2
)

replace github.com/omec-project/upf-epc v0.3.0 => github.com/ardzoht/omec-upf v0.6.0
