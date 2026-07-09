module github.com/burgrp/din-rs485-wifi/fw

go 1.25.2

require github.com/burgrp/bleriot/lib v0.0.0-00010101000000-000000000000

require (
	github.com/burgrp/reg v1.0.12 // indirect
	github.com/burgrp/tinygo-drivers/bb/spi v1.0.0 // indirect
	github.com/burgrp/tinygo-drivers/pan211x v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/lmittmann/tint v1.1.3 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
)

replace github.com/burgrp/bleriot/lib => ../../bleriot/lib
