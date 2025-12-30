import asyncio

class AnalyticsEngine:
    def __init__(self):
        pass

    async def run(self):
        print('Hello World!')

if __name__ == '__main__':
    engine = AnalyticsEngine()
    asyncio.run(engine.run())