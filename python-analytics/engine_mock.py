'''Mock engine to test Python TUI client'''
import asyncio
import json
import random
import datetime

async def handle_client(reader, writer):
    print("Client connected to Mock Engine...")
    try:
        while True:
            # Generate fake trade
            now = datetime.datetime.now().strftime("%H:%M:%S.%f")[:-3]
            trade = {
                "time": now,
                "symbol": random.choice(["AAPL", "GOOGL", "MSFT", "TSLA"]),
                "price": round(random.uniform(150.0, 300.0), 2),
                "vwap": round(random.uniform(150.0, 300.0), 2)
            }
            
            # Send newline-delimited JSON
            msg = json.dumps(trade) + "\n"
            writer.write(msg.encode('utf-8'))
            await writer.drain()
            
            # Random delay to simulate market speed
            await asyncio.sleep(random.uniform(0.1, 0.5))
    except Exception:
        pass
    finally:
        print("Client disconnected.")

async def main():
    server = await asyncio.start_server(handle_client, 'localhost', 8888)
    print("Mock Engine running on :8888...")
    async with server:
        await server.serve_forever()

if __name__ == "__main__":
    asyncio.run(main())
