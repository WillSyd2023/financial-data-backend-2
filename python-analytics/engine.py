'''Real-time Python financial analytics engine'''
import asyncio
import sys
import logging
import yaml
from aiokafka import AIOKafkaConsumer

logging.basicConfig(level=logging.INFO, format='%(levelname)s: %(message)s')

class AnalyticsEngine:
    def __init__(self):
        pass

    async def run(self):
        # Load config, but just if fail
        try:
            with open('../config/config.yml', 'r', encoding='utf-8') as f:
                config = yaml.safe_load(f)
        except yaml.YAMLError as exc:
            logging.critical('Failed to load config: %s', exc)
            sys.exit(1)

        # Setup Kafka Consumer
        consumer = AIOKafkaConsumer(
            config['kafka']['topic'],
            bootstrap_servers=config['kafka']['broker_url'],
            group_id='analytics-engine-group'
        )

if __name__ == '__main__':
    engine = AnalyticsEngine()
    asyncio.run(engine.run())