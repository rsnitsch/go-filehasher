#!/usr/bin/env python3
"""
Generate files filled with random data.

Copyright (C) 2012-2013 Robert Nitsch
Licensed according to GPL v3.
"""
import argparse
import os
import random
import sys

def write_random_data(file, size, progress_callback=None):
    for i in range(size):
        file.write(bytes([random.randint(0, 255)]))

        if progress_callback is not None and i % 2**18 == 0:
            progress_callback(i)

def main(argv):
    parser = argparse.ArgumentParser(description="Generate files filled with random data")
    parser.add_argument("target", help="The target file")
    parser.add_argument("size", type=int, help="The target size")
    parser.add_argument("--force", "--overwrite", action="store_true", default=False, help="Overwrite existing files")
    args = parser.parse_args()

    if os.path.isdir(args.target):
        print("Target is a directory: '%s'" % args.target, file=sys.stderr)
        return 1

    if not args.force and os.path.isfile(args.target):
        print("Target file does already exist: '%s'. Add --force to overwrite." % args.target, file=sys.stderr)
        return 1

    if args.size < 0:
        print("Size must be non-negative.", file=sys.stderr)
        return 1

    with open(args.target, "w+b") as fh:
        write_random_data(fh, args.size, lambda i: print("\r%d / %d done (%.2f%%)" % (i, args.size, float(i)/args.size*100), end=""))

    print("\rTarget file generation has completed.")

if __name__ == "__main__":
    sys.exit(main(sys.argv))
