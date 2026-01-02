import asyncio
import json
from rich.live import Live
from rich.table import Table

def generate_table(rows):
    table = Table(title="Real-Time VWAP Monitor (TCP Stream)")
    table.add_column("Time", style="cyan")
    table.add_column("Symbol", style="magenta")
    table.add_column("Price", justify="right", style="green")
    table.add_column("VWAP (Rolling)", justify="right", style="yellow")
    
    for row in rows:
        table.add_row(row['time'], row['symbol'], f"${row['price']:.2f}", f"${row['vwap']:.2f}")
    return table

async def simple_client():
    addr = 'Unknown'
    history = [] 
    
    try:
        reader, writer = await asyncio.open_connection('localhost', 8888)
        addr = writer.get_extra_info('sockname')
        print(f"{addr} is connected to Engine... (Ctrl+C to stop)")

        with Live(generate_table(history), refresh_per_second=4) as live:
            while True:
                line = await reader.readline()
                if not line:
                    break
                try:
                    data = json.loads(line)
                    history.insert(0, data)
                    if len(history) > 15:
                        history.pop()
                    
                    # Update the UI
                    live.update(generate_table(history))
                    
                except json.JSONDecodeError:
                    pass
    except Exception as e:
        print(f"Error: {e}")
    finally:
        print(f"{addr} disconnected")

if __name__ == "__main__":
    try:
        asyncio.run(simple_client())
    except KeyboardInterrupt:
        pass