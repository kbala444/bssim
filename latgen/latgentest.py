import unittest
import latgen as lg
import os

class LatgenTest(unittest.TestCase):
    def test_get_num_nodes(self):
        with open('testwl', 'w+') as f:
            f.write('node_count:100, latency:0, other:10')

        self.assertEqual(lg.get_num_nodes('testwl'), 100)
        with open('testwl', 'w') as f:
            f.write('latency:0,  node_count: 5,  other:10')

        self.assertEqual(lg.get_num_nodes('testwl'), 5)

    def test_set_manual(self):
        with open('testwl', 'w+') as f:
            f.write('node_count:100, latency:0, other:10')

        lg.set_manual('testwl')
        with open('testwl', 'r') as f:
            self.assertTrue('manual_links: true' in f.readline())

        lg.set_manual('testwl')
        # ensure more 'manual_links' are not added
        with open('testwl', 'r') as f:
            config = f.readline()
            self.assertTrue(config.count('manual') == 1)

    def tearDown(self):
        os.remove('testwl')

if __name__ == '__main__':
    unittest.main()
    cleanup()
