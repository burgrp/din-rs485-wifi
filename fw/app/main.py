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

print('MODBUS Energy Meter to WiFi bridge')

class Register:
    def __init__(self, device, regdef):
        self.device = device
        self.regdef = regdef
        self.value = None

class RegistryHandler:

    registers = {}

    def __init__(self):
        for device in site_config.devices:

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
                self.registers[name] = Register(device, regdef)

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


    def get_names(self):
        return self.registers.keys()

    def get_meta(self, name):
        return {
            'device': 'test',
            'title': self.registers[name].regdef.title,
            'type': 'number',
            'unit': self.registers[name].regdef.unit
        }

    def get_value(self, name):
        return self.registers[name].value

    def set_value(self, name, value):
        self.registers[name].value = value

registryHandler = RegistryHandler()

registry = mqtt_reg.Registry(
    registryHandler,
    wifi_ssid=site_config.wifi_ssid,
    wifi_password=site_config.wifi_password,
    mqtt_broker=site_config.mqtt_broker,
    ledPin=4,
    debug=site_config.debug
)

registry.start()

txEn = Pin(3, Pin.OUT)

def serial_mode(mode):
    if mode == modbus_rtu.serial_cb_tx_begin:
        txEn.on()
    elif mode == modbus_rtu.serial_cb_tx_end:
        time.sleep(0.01)
        txEn.off()

while True:
    for device in site_config.devices:

        if site_config.debug:
            print('Device:', device['address'], device['name'])

        def set_value(regdef, value):
            name = device['name'] + '.' + regdef.name
            if site_config.debug:
                print(name, '=', value)
            registryHandler.registers[name].value = value

        if not site_config.emu:
            uart = UART(1, baudrate=device['baud'], tx=21, rx=20, timeout=1000, timeout_char=1000)
            master = modbus_rtu.RtuMaster(uart, serial_mode)

        for group in device['groups']:
            first_address = group[0].address
            last_address = group[-1].address

            try:
                words = None
                word_count = 2 * (last_address - first_address + 1)

                if not site_config.emu:
                    words = master.execute(device['address'], modbus.defines.READ_INPUT_REGISTERS, first_address, word_count)
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


        if not site_config.emu:
            uart.deinit()

        time.sleep(1)