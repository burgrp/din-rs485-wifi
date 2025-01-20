from machine import UART, Pin
from mb import ModbusRTUClient
import time
import struct
import mqtt_reg
import asyncio

import micropython, gc

import sys
sys.path.append('/')

import site_config
import device_config

async def main():

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

    registry.start()

    master = ModbusRTUClient()

    # while True:

    #     micropython.mem_info()
    #     print("Alloc:", gc.mem_alloc(), "Free:", gc.mem_free())

    #     for device in device_config.devices:

    #         if device_config.debug:
    #             print('Device:', device['address'], device['name'])

    #         def set_value(regdef, value):
    #             name = device['name'] + '.' + regdef.name
    #             if device_config.debug:
    #                 print(name, '=', value)
    #             registers[name].set_value_local(value)


    #         for group in device['groups']:
    #             first_address = group[0].address
    #             last_address = group[-1].address

    #             try:
    #                 data = None

    #                 attempt = 0
    #                 while True:
    #                     attempt += 1
    #                     try:
    #                         #data = await master.read_holding_registers(device['address'], start_addr=first_address, count=last_address - first_address + 2)
    #                         break
    #                     except Exception as e:
    #                         if attempt >= 5:
    #                             raise e
    #                         print('error reading group data, retrying...')
    #                         await asyncio.sleep(1)

    #                 for regdef in group:
    #                     offset = regdef.address - first_address
    #                     value = struct.unpack('<f', struct.pack('<h', int(data[offset + 1])) + struct.pack('<h', int(data[offset + 0])))[0]
    #                     set_value(regdef, value)

    #             except Exception as e:
    #                 print('error reading group data:', e)
    #                 for regdef in group:
    #                     set_value(regdef, None)
    #                 await asyncio.sleep(1)

    #         await asyncio.sleep(1)

    await asyncio.sleep(10000)

asyncio.run(main())
