# modbus_async.py
import uasyncio as asyncio
from machine import UART, Pin
import time

# -------------------------------------------------
# Configuration
# -------------------------------------------------
MODBUS_BAUDRATE = 9600
MODBUS_UART_ID = 1    # e.g., on ESP32-C3
MODBUS_DE_PIN = 3    # Example driver-enable pin
# Adjust these as needed:
TX_PIN = 21
RX_PIN = 20

# -------------------------------------------------
# Simple helper to manage RS485 driver-enable line
# -------------------------------------------------
class RS485:
    def __init__(self, de_pin: int):
        self._de = Pin(de_pin, Pin.OUT)
        self._de.value(0)  # Disable driver by default

    def enable_tx(self):
        self._de.value(1)

    def enable_rx(self):
        self._de.value(0)

# -------------------------------------------------
# Modbus RTU Client Class
# -------------------------------------------------
class ModbusRTUClient:
    def __init__(self, uart_id=MODBUS_UART_ID, baudrate=MODBUS_BAUDRATE,
                 tx_pin=TX_PIN, rx_pin=RX_PIN, de_pin=MODBUS_DE_PIN):
        self.rs485 = RS485(de_pin)
        self.uart = UART(uart_id, baudrate=baudrate, tx=Pin(tx_pin), rx=Pin(rx_pin))

    def _calc_crc(self, data: bytes) -> int:
        """Compute Modbus CRC-16."""
        crc = 0xFFFF
        for b in data:
            crc ^= b
            for _ in range(8):
                lsb = crc & 0x0001
                crc >>= 1
                if lsb:
                    crc ^= 0xA001
        return crc

    async def _send_request(self, packet: bytes):
        """Send bytes over UART with RS485 driver enable."""
        self.rs485.enable_tx()
        await asyncio.sleep_ms(2)  # short delay to allow DE to settle
        self.uart.write(packet)
        # Wait for packet transmission to finish
        # (Some chips need a flush or a fixed delay)
        await asyncio.sleep_ms(int((len(packet) * 11 * 1000) / MODBUS_BAUDRATE))
        self.rs485.enable_rx()

    async def _receive_response(self, expected_len: int, timeout_ms=100):
        """Read expected bytes from UART with a timeout."""
        start = time.ticks_ms()
        received = b""
        while time.ticks_diff(time.ticks_ms(), start) < timeout_ms:
            if self.uart.any():
                chunk = self.uart.read()
                if chunk:
                    received += chunk
            await asyncio.sleep_ms(1)
            if len(received) >= expected_len:
                break
        return received

    async def read_holding_registers(self, device_id: int, start_addr: int, count: int):
        """Read multiple holding registers using function code 0x03."""
        function_code = 0x03
        # Build request: [device_id, function_code, start_addr_hi, start_addr_lo, count_hi, count_lo, CRC16(2 bytes)]
        request = bytearray([
            device_id,
            function_code,
            (start_addr >> 8) & 0xFF,
            start_addr & 0xFF,
            (count >> 8) & 0xFF,
            count & 0xFF
        ])
        crc = self._calc_crc(request)
        request += bytearray([crc & 0xFF, (crc >> 8) & 0xFF])

        # Send request
        await self._send_request(request)

        # Expected response: device_id, function_code, byte_count, data..., CRC16
        # byte_count should be count * 2
        # So total length = 5 (id + func + byte_count) + count*2 + 2 CRC = 5 + count*2 + 2
        expected_len = 5 + (count * 2) + 2
        response = await self._receive_response(expected_len)
        return response

    async def write_single_register(self, device_id: int, reg_addr: int, value: int):
        """Write a single holding register using function code 0x06."""
        function_code = 0x06
        request = bytearray([
            device_id,
            function_code,
            (reg_addr >> 8) & 0xFF,
            reg_addr & 0xFF,
            (value >> 8) & 0xFF,
            value & 0xFF
        ])
        crc = self._calc_crc(request)
        request += bytearray([crc & 0xFF, (crc >> 8) & 0xFF])

        await self._send_request(request)
        # Response is echo of request: 8 bytes total
        response = await self._receive_response(expected_len=8)
        return response

# -------------------------------------------------
# Example usage
# -------------------------------------------------
# async def main():
#     modbus = ModbusRTUClient()
#     while True:
#         # Read holding registers example
#         resp = await modbus.read_holding_registers(device_id=1, start_addr=0x0000, count=2)
#         print("Read registers response:", resp)

#         # Write single register example
#         resp = await modbus.write_single_register(device_id=1, reg_addr=0x0001, value=123)
#         print("Write single register response:", resp)

#         await asyncio.sleep(2)

# -------------------------------------------------
# Run main if this is the entry point
# -------------------------------------------------
# import uasyncio
# uasyncio.run(main())
