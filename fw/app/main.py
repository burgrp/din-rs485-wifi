# from machine import UART
# from modbus import modbus_rtu
# import modbus
import time
# import struct
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
        self.value = 1

class RegistryHandler:

    registers = {}

    def __init__(self):
        for device in site_config.devices:
            for regdef in device["registers"]:
                name = device["name"] + "." + regdef.name
                self.registers[name] = Register(device, regdef)

    def get_names(self):
        return self.registers.keys()

    def get_meta(self, name):
        return {
            "device": "test",
            "title": self.registers[name].regdef.title,
            "type": "number",
            "unit": self.registers[name].regdef.unit
        }

    def get_value(self, name):
        return self.registers[name].value

    def set_value(self, name, value):
        self.registers[name].value = value


registry = mqtt_reg.Registry(
    RegistryHandler(),
    wifi_ssid=site_config.wifi_ssid,
    wifi_password=site_config.wifi_password,
    mqtt_broker=site_config.mqtt_broker,
    ledPin=4,
    debug=site_config.debug
)

registry.start()

# txEn = Pin(3, Pin.OUT)

# def serial_mode(mode):
#     if mode == modbus_rtu.serial_cb_tx_begin:
#         led.on()
#         txEn.on()
#     elif mode == modbus_rtu.serial_cb_tx_end:
#         time.sleep(0.01)
#         led.off()
#         txEn.off()

# while True:
#     for device in site_config.devices:
#         print('Device:', device['address'], device['name'])

#         uart = UART(1, baudrate=device['baud'], tx=21, rx=20, timeout=1000, timeout_char=1000)
#         master = modbus_rtu.RtuMaster(uart, serial_mode)

#         for register in registers:
#             try:
#                 if site_config.emu:
#                     register.set(1)
#                     time.sleep(0.1)
#                 else:
#                     f_word_pair = master.execute(device['address'], modbus.defines.READ_INPUT_REGISTERS, register.address, 2)
#                     register.set(struct.unpack('<f', struct.pack('<h', int(f_word_pair[1])) + struct.pack('<h', int(f_word_pair[0])))[0])

#             except Exception as e:
#                 register.set(None)
#                 print('error:', e)

#             time.sleep(.1)

#         uart.deinit()
