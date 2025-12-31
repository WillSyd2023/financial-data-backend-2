import asyncio
import json

async def simple_client():
    addr = 'Unknown'
    try:
        reader, writer = await asyncio.open_connection('localhost', 8888)
        addr = writer.get_extra_info('sockname')
        print(f"{addr} is connected to Engine... (Ctrl+C to stop)")
        while True:
            line = await reader.readline()
            if not line:
                break
            try: # Parse and Print
                data = json.loads(line)
                time = data['time']
                sym = data['symbol']
                price = data['price']
                vwap = data['vwap']
                print(
                    f"[{time}] {sym} | Price: ${price:.2f} | VWAP: ${vwap:.2f}")
            except json.JSONDecodeError:
                print(f"Raw: {line}")
    except Exception as e:
        print(f"Error: {e}")
    finally:
        print(f"{addr} disconnected")

if __name__ == "__main__":
    try:
        asyncio.run(simple_client())
    except KeyboardInterrupt:
        pass