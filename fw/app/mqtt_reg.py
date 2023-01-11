import mqtt_as
import uasyncio
import _thread
import ujson
import uio
from machine import Pin


class Registry:

    def __init__(self, handler, wifi_ssid, wifi_password, mqtt_broker, ledPin=2, ledLogic=True, debug=False):
        self.handler = handler
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

    async def publish_register_value(self, name):
        value = await self.handler.get_value(name)
        await self.__publish_json('register/'+name+'/is', value)

    async def advertise_registers(self):
        for name in self.handler.get_names():
            meta = self.handler.get_meta(name)
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

                for name in self.handler.get_names():
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
                try:
                    topic = topic.decode()
                    print(topic, "-->")

                    if topic == 'register/advertise!':
                        await self.advertise_registers()

                    else:
                        value = None if len(message) == 0 else ujson.load(
                            uio.BytesIO(message.decode()))

                        for name in self.handler.get_names():

                            if topic == 'register/'+name+'/get':
                                await self.publish_register_value(name)

                            if topic == 'register/'+name+'/set':
                                await self.handler.set_value(name, value)
                                await self.publish_register_value(name)

                    print("<--", topic)

                except Exception as e:
                    if self.debug:
                        print('Error handling message because:', e)

        await read_messages()

    def start(self):
        _thread.stack_size(32768)
        _thread.start_new_thread(lambda: uasyncio.run(self.run()), ())


# class TestHandler:
#     registers = {
#         "a": 1,
#         "b": 2,
#         "c": 3
#     }

#     def get_names(self):
#         return self.registers.keys()

#     def get_meta(self, name):
#         return {
#             "device": "test",
#             "title": "test register " + name
#         }

#     async def get_value(self, name):
#         return self.registers[name]

#     async def set_value(self, name, value):
#         self.registers[name] = value


# registry = mqtt_reg.Registry(
#     TestHandler(),
#     wifi_ssid=site_config.wifi_ssid,
#     wifi_password=site_config.wifi_password,
#     mqtt_broker=site_config.mqtt_broker,
#     ledPin=4,
#     debug=site_config.debug
# )

# registry.start()
