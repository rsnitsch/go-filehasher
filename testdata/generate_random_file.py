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
        for i in range(args.size):
            fh.write(bytes([random.randint(0, 255)]))

            if i % 2**18 == 0:
                print("\r%d / %d done (%.2f%%)" % (i, args.size, float(i)/args.size*100), end="")
    print("\rTarget file generation has completed.")

if __name__ == "__main__":
    sys.exit(main(sys.argv))
