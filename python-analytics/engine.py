'''Real-time Python financial analytics engine'''
import asyncio
import sys
import logging
import json
from collections import deque
import yaml
from aiokafka import AIOKafkaConsumer, errors

logging.basicConfig(level=logging.INFO, format='%(levelname)s: %(message)s')

# O(1) Analytics Logic
class MarketMetrics:
    def __init__(self, window_size=50):
        self.window_size = window_size
        self.prices_and_volumes = deque()
        self.cum_price_volume = 0.0
        self.cum_volume = 0.0
        
    def update(self, price):
        pass

class AnalyticsEngine:
    def __init__(self):
        # Map: Symbol -> MarketMetrics
        self.metrics = {}

    async def run(self):
        # Load config, but just if fail
        try:
            with open('../config/config.yml', 'r', encoding='utf-8') as f:
                config = yaml.safe_load(f)
        except yaml.YAMLError as exc:
            logging.critical('Failed to load config: %s', exc)
            sys.exit(1)

        # Setup Kafka Consumer, then wait for Kafka to be ready
        consumer = AIOKafkaConsumer(
            config['kafka']['topic'],
            bootstrap_servers=config['kafka']['broker_url'],
            group_id='analytics-engine-group'
        )
        while True:
            try:
                await consumer.start()
                break
            except errors.KafkaConnectionError:
                logging.warning('Kafka not ready yet. Retrying in 2 seconds...')
                await asyncio.sleep(2)
        print('''Kafka reader configured successfully. 
	            Consumer Group ID: analytics-engine-group''')

        try:
            async for msg in consumer:
                print('Message received | Topic: %s | Partition: %d | Offset: %d',
			        msg.topic, msg.partition, msg.offset)

                value = json.loads(msg.value.decode('utf-8'))
                print('Message value:', value)

                
        finally:
            await consumer.stop()

if __name__ == '__main__':
    engine = AnalyticsEngine()
    asyncio.run(engine.run())