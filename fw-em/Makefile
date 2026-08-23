ifndef BLERIOT_MAKE

$(or $(MAKECMDGOALS),build): __bleriot_bounce
	@:
.PHONY: __bleriot_bounce $(MAKECMDGOALS)
__bleriot_bounce:
	@go run . make $(MAKECMDGOALS)

else

IMAGE = image.elf

.PHONY: build flash gdb rtt install-pack disassembly

build:
	tinygo build --stack-size 1024b --size html --serial rtt --scheduler tasks -print-allocs=. --gc leaking --target ${TARGET_TINYGO} -o ${IMAGE} .

flash: build
	pkill pyocd || true
	sleep 1
	pyocd load -t ${TARGET_PYOCD} ${IMAGE}
	pyocd rtt -t ${TARGET_PYOCD}

gdb:
	pyocd gdb --target ${TARGET_PYOCD} --semihosting --persist

rtt:
	pyocd rtt -t ${TARGET_PYOCD}

install-pack:
	pyocd pack install ${CMSIS_PACK}

disassembly: build
	arm-none-eabi-objdump -dS ${IMAGE} > disassembly.txt

endif