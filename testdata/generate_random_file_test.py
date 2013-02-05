#!/usr/bin/env python3
"""
Unittests for generate_random_file.py.

Copyright (C) 2012-2013 Robert Nitsch
Licensed according to GPL v3.
"""
import io
import unittest

from generate_random_file import write_random_data

class TestGenerateRandomFile(unittest.TestCase):
    def test_size(self):
        for size in [0, 11, 17, 1024*512 + 13]:
            file = io.BytesIO()
            write_random_data(file, size)
            self.assertEqual(len(file.getbuffer()), size)

if __name__ == "__main__":
    unittest.main()
