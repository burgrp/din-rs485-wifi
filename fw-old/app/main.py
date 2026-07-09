from machine import UART, Pin
from modbus import modbus_rtu
import modbus
import time
import struct
import mqtt_reg
import uasyncio

import sys
sys.path.append('/')

import site_config
import device_config

print('MODBUS Energy Meter to WiFi bridge')

class ModbusRegister(mqtt_reg.ServerReadOnlyRegister):
    def __init__(self, device, regdef):
        super().__init__(
            device['name'] + '.' + regdef.name,
            {
                    'device': device['name'],
                    'title': regdef.title,
                    'type': 'number',
                    'unit': regdef.unit
            }
        )
        self.value = None

registers = {}

for device in device_config.devices:

    print('Device', device['name'], 'at address', device['address'])

    groups = []
    group = []

    def flush_group():
        if len(group) > 0:
            groups.append(group.copy())
            group.clear()

    last_addr = None
    for regdef in device['registers']:
        name = device['name'] + '.' + regdef.name
        registers[name] = ModbusRegister(device, regdef)

        group.append(regdef)

        if last_addr != None and regdef.address > last_addr + 4:
            flush_group()

        last_addr = regdef.address

    flush_group()

    device['groups'] = groups

    for group in groups:
        print(" group")
        for regdef in group:
            print('  ', regdef.address, regdef.name, '"' + regdef.title + '"', regdef.unit)

led = Pin(4, Pin.OUT)

registry = mqtt_reg.Registry(
    wifi_ssid=site_config.wifi_ssid,
    wifi_password=site_config.wifi_password,
    mqtt_broker=site_config.mqtt_broker,
    server=list(registers.values()),
    online_cb=lambda online: led.value(online),
    debug=device_config.debug
)

registry.start(background=True)

txEn = Pin(3, Pin.OUT)

def serial_mode(mode):
    if mode == modbus_rtu.serial_cb_tx_begin:
        txEn.on()
    elif mode == modbus_rtu.serial_cb_tx_end:
        time.sleep(0.01)
        txEn.off()

while True:
    for device in device_config.devices:

        if device_config.debug:
            print('Device:', device['address'], device['name'])

        def set_value(regdef, value):
            name = device['name'] + '.' + regdef.name
            if device_config.debug:
                print(name, '=', value)
            registers[name].set_value_local(value)

        if not device_config.emu:
            uart = UART(1, baudrate=device['baud'], tx=21, rx=20, timeout=1000, timeout_char=200)
            master = modbus_rtu.RtuMaster(uart, serial_mode)

        for group in device['groups']:
            first_address = group[0].address
            last_address = group[-1].address

            try:
                words = None
                word_count = 2 * (last_address - first_address + 1)

                if not device_config.emu:

                    attempt = 0
                    while True:
                        attempt += 1
                        try:
                            words = master.execute(device['address'], modbus.defines.READ_INPUT_REGISTERS, first_address, word_count)
                            break
                        except Exception as e:
                            if attempt >= 5:
                                raise e
                            print('error reading group data, retrying...')
                            time.sleep(1)

                else:
                    words = []
                    for i in range(word_count):
                        words.append(1)
                    time.sleep(0.1)

                for regdef in group:
                    offset = regdef.address - first_address
                    value = struct.unpack('<f', struct.pack('<h', int(words[offset + 1])) + struct.pack('<h', int(words[offset + 0])))[0]
                    set_value(regdef, value)

            except Exception as e:
                print('error reading group data:', e)
                for regdef in group:
                    set_value(regdef, None)
                time.sleep(1)


        if not device_config.emu:
            uart.deinit()
            uart = None
            master = None

        time.sleep(1)