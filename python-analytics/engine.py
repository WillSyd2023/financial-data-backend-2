'''Real-time Python financial analytics engine'''
import asyncio
import sys
import logging
import json
from collections import deque, defaultdict
import yaml
from aiokafka import AIOKafkaConsumer, errors

logging.basicConfig(level=logging.INFO, format='%(levelname)s: %(message)s')

# O(1) Analytics Logic
class MarketMetrics:
    def __init__(self, window_size=50):
        self.window_size = window_size
        self.pvs = deque()
        self.cum_pv = 0.0
        self.cum_vol = 0
        
    def update(self, data):
        '''Calculate new "rolling volume-weight average price"'''
        pv = {'p': data['p'], 'v': data['v']}
        if len(self.pvs) == self.window_size:
            prev = self.pvs.popleft()
            self.cum_pv -= prev['p'] * prev['v']
            self.cum_vol -= prev['v']
        self.pvs.append(pv)
        self.cum_pv += pv['p'] * pv['v']
        self.cum_vol += pv['v']
        if self.cum_vol == 0:
            return None
        return self.cum_pv/self.cum_vol

class AnalyticsEngine:
    def __init__(self):
        # Set of active TCP writers
        self.clients = set()

        # Map: Symbol -> MarketMetrics
        self.metrics = defaultdict(lambda: MarketMetrics())

    async def send(self, data, w):
        '''Broadcast most-recently-updated symbol information to a single client'''
        try:
            w.write(data)
            await w.drain()
        except:
            pass


    async def broadcast(self, time, symbol, price, vwap):
        '''Broadcast most-recently-updated symbol information to clients'''
        data = (json.dumps({
            'time': time,
            'symbol': symbol,
            'price': price,
            'vwap': vwap,
        }) + '\n').encode('utf-8')

        if self.clients:
            await asyncio.gather(*(self.send(data, w) for w in list(self.clients)))

    async def handle_client(self, reader, writer):
        '''TCP client handler callback function'''
        addr = writer.get_extra_info('peername')
        print(f'New Client Connection: {addr}')
        self.clients.add(writer)
        try:
            while True:
                data = await reader.read()
                if not data:
                    break
        except asyncio.CancelledError:
            pass
        finally:
            print(f'Client Disconnected: {addr}')
            self.clients.discard(writer)
            writer.close()
            await writer.wait_closed()

    async def run(self):
        '''The actually start of the engine's operation'''
        # Load config, but just quit if fail
        try:
            with open('../config/config.yml', 'r', encoding='utf-8') as f:
                config = yaml.safe_load(f)
        except yaml.YAMLError as exc:
            logging.critical('Failed to load config: %s', exc)
            sys.exit(1)

        # Start TCP Server
        host = config['analytics_engine']['tcp_host']
        port = config['analytics_engine']['tcp_port']
        await asyncio.start_server(self.handle_client, host, port)
        print(f'TCP Server listening on port {port}')

        # Setup Kafka Consumer, then wait for Kafka to be ready
        consumer = AIOKafkaConsumer(
            config['kafka']['topic'],
            bootstrap_servers=config['kafka']['broker_url'],
            group_id='analytics-engine-group'
        )
        while True:
            try:
                await consumer.start(1024)
                break
            except errors.KafkaConnectionError:
                logging.warning('Kafka not ready yet. Retrying in 2 seconds...')
                await asyncio.sleep(2)
        print(
        'Kafka reader configured successfully. Consumer Group ID: analytics-engine-group')

        try:
            async for msg in consumer:
                try:
                    print('Message received | Topic: %s | Partition: %d | Offset: %d',
                        msg.topic, msg.partition, msg.offset)

                    value = json.loads(msg.value.decode('utf-8'))
                    print('Message value:', value)

                    for data in value['data']:
                        vwap = self.metrics[data['s']].update(data)
                        await self.broadcast(data['t'], data['s'], data['p'], vwap)
                        print(f"Processed {data['s']}: ${data['p']} at time {data['t']} | VWAP: ${vwap}")
                except Exception as e:
                    print(f'Error processing message: {e}')
        finally:
            await consumer.stop()

if __name__ == '__main__':
    engine = AnalyticsEngine()
    asyncio.run(engine.run())