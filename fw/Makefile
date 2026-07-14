# RS485 node firmware. One firmware per device, sharing the uartline transport
# and rig runtime. Select which one with CMD (meter or bms).
# Mirrors the BleRiot example Makefile style.

# TARGET_TINYGO ?= py32f030_64k_8k
# TARGET_PYOCD  ?= py32f030x8
# CMSIS_PACK    ?= PY32F030

TARGET_TINYGO = py32f003_32k_4k
TARGET_PYOCD = py32f003x6
CMSIS_PACK = PY32F003

# Firmware to build/flash/provision: meter | bms
CMD ?= meter

IMAGE = ${CMD}.elf

.PHONY: build flash gdb rtt install-pack disassembly new provision hub test

build:
	tinygo build --stack-size 1024b --size html --serial uart --scheduler tasks -print-allocs=. --gc leaking --target ${TARGET_TINYGO} -o ${IMAGE} ./${CMD}

flash: build
	pkill pyocd || true
	sleep 1
	pyocd load -t ${TARGET_PYOCD} ${IMAGE}

gdb:
	pyocd gdb --target ${TARGET_PYOCD} --semihosting --persist

rtt:
	pyocd rtt -t ${TARGET_PYOCD}

install-pack:
	pyocd pack install ${CMSIS_PACK}

disassembly: build
	arm-none-eabi-objdump -dS ${IMAGE} > disassembly.txt

new:
	go run . new

provision:
	go run . provision

hub:
	go run . hub --diagnostics rf

test:
	go test ./...
