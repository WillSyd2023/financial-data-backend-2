import asyncio
import json

async def simple_client():
    addr = 'Unknown'
    try:
        reader, writer = await asyncio.open_connection('0.0.0.0', 8888)
        addr = writer.get_extra_info('sockname')
        print(f"{addr} is connected to Engine... (Ctrl+C to stop)")
        while True:
            line = await reader.readline()
            if not line:
                break
                
            # Parse and Print
            try:
                data = json.loads(line)
                print(f"[{data['time']}] {data['symbol']} | Price: \
                    ${data['price']:.2f} | VWAP: ${data['vwap']:.2f}")
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