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

# class Register:
#     def __init__(self, definition):
#         self.definition = definition
#         self.value = None

#     def get_name(self):
#         return self.name

#     def get_meta(self):
#         return {
#             "device": "test"
#         }

#     async def get_value(self):
#         return self.value

#     async def set_value(self, value):
#         self.value = value


# class Device:
#     registers = []

# devices = []

# for device in site_config.devices:
#     for r in devices.registers:

registers = {
    "a": 1,
    "b": 2,
    "c": 3
}

registry = mqtt_reg.Registry(
    registers,
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
