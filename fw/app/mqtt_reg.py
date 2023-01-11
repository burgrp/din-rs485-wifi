import mqtt_as
import uasyncio
import _thread
import ujson
import uio
from machine import Pin


class Registry:

    def __init__(self, registers, wifi_ssid, wifi_password, mqtt_broker, ledPin=2, ledLogic=True, debug=False):
        self.registers = registers
        self.debug = debug
        mqtt_config = mqtt_as.config.copy()
        mqtt_config['ssid'] = wifi_ssid
        mqtt_config['wifi_pw'] = wifi_password
        mqtt_config['server'] = mqtt_broker
        mqtt_config['queue_len'] = 1

        mqtt_as.MQTTClient.DEBUG = debug
        self.mqtt_client = mqtt_as.MQTTClient(mqtt_config)

        ledPin = Pin(ledPin, Pin.OUT)
        self.led = lambda on: ledPin.value(on == ledLogic)

        self.led(False)

    async def __publish_json(self, topic, val):
        message = uio.BytesIO()

        if val != None:
            ujson.dump(val, message)

        message = message.getvalue()

        await self.mqtt_client.publish(topic, message)

    async def publish_register_value(self, register):
        name = register.get_name()
        value = await register.get_value()
        await self.__publish_json('register/'+name+'/is', value)

    async def advertise_registers(self):
        for register in self.registers:
            name = register.get_name()
            meta = register.get_meta()
            await self.__publish_json('register/'+name+'/advertise', meta)

    async def run(self):
        await self.mqtt_client.connect()

        async def up_event_loop():
            while True:
                await self.mqtt_client.up.wait()
                self.mqtt_client.up.clear()
                self.led(True)

                async def subscribe(topic):
                    if self.debug:
                        print('Subscribing to:', topic)

                    await self.mqtt_client.subscribe(topic, 1)

                await subscribe('register/advertise!')

                for register in self.registers:
                    name = register.get_name()
                    await subscribe('register/'+name+'/get')
                    await subscribe('register/'+name+'/set')

                await self.advertise_registers()

        uasyncio.create_task(up_event_loop())

        async def down_event_loop():
            while True:
                await self.mqtt_client.down.wait()
                self.mqtt_client.down.clear()
                self.led(False)

        uasyncio.create_task(down_event_loop())

        async def read_messages():
            async for topic, message, retained in self.mqtt_client.queue:
                if not retained:
                    try:
                        topic = topic.decode()
                        if topic == 'register/advertise!':
                            await self.advertise_registers()

                        else:
                            value = None if len(message) == 0 else ujson.load(
                                uio.BytesIO(message.decode()))

                            for register in self.registers:
                                name = register.get_name()

                                if topic == 'register/'+name+'/get':
                                    await self.publish_register_value(register)

                                if topic == 'register/'+name+'/set':
                                    await register.set_value(value)
                                    await self.publish_register_value(register)

                    except Exception as e:
                        if self.debug:
                            print('Error handling message because:', e)

        await read_messages()

    def start(self):
        _thread.stack_size(32768)
        _thread.start_new_thread(lambda: uasyncio.run(self.run()), ())


# class TestRegister:

#     def __init__(self, name):
#         self.name = name
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


# testRegisters = (
#     TestRegister('a'),
#     TestRegister('b')
# )
