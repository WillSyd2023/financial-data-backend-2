import unittest
from engine import MarketMetrics

class TestVWAP(unittest.TestCase):
    def test_basic_vwap(self):
        metrics = MarketMetrics(window_size=2)
        self.assertAlmostEqual(metrics.update({'p': 100, 'v': 1}), 100.0)
        self.assertAlmostEqual(metrics.update({'p': 200, 'v': 1}), 150.0)

    def test_rolling_eviction(self):
        metrics = MarketMetrics(window_size=2)
        metrics.update({'p': 100, 'v': 1})
        metrics.update({'p': 200, 'v': 1})
        self.assertAlmostEqual(metrics.update({'p': 300, 'v': 1}), 250.0)

    def test_zero_volume_safety(self):
        metrics = MarketMetrics()
        result = metrics.update({'p': 100, 'v': 0})
        self.assertIsNone(result)

if __name__ == '__main__':
    unittest.main()