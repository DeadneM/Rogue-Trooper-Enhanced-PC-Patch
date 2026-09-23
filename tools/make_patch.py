#!/usr/bin/env python3
"""Build the verified RTDP1 patch payload used by the Rogue Trooper patcher.

Usage:
    python tools/make_patch.py ORIGINAL.exe V64.exe patcher/rogue_v64.rtdp1.zlib

The generated payload is source-locked by SHA-256 and contains a verified
instruction stream. The current V64 payload uses a literal target instruction
because the retail executable is packed while V64 is an unpacked/rebuilt PE,
so copy-based binary deltas provide little benefit.
"""
from __future__ import annotations

import hashlib
import struct
import sys
import zlib
from pathlib import Path

MAGIC = b"RTDP1\0"


def main() -> int:
    if len(sys.argv) != 4:
        print("Usage: make_patch.py ORIGINAL.exe V64.exe OUTPUT.zlib")
        return 2

    source_path, target_path, output_path = map(Path, sys.argv[1:])
    source = source_path.read_bytes()
    target = target_path.read_bytes()

    raw = bytearray(MAGIC)
    raw += struct.pack("<QQ", len(source), len(target))
    raw += hashlib.sha256(source).digest()
    raw += hashlib.sha256(target).digest()
    raw += struct.pack("<I", 1)
    raw += b"\x01" + struct.pack("<I", len(target)) + target

    compressed = zlib.compress(bytes(raw), 9)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_bytes(compressed)

    print("source sha256:", hashlib.sha256(source).hexdigest())
    print("target sha256:", hashlib.sha256(target).hexdigest())
    print("payload bytes:", len(compressed))
    print("payload sha256:", hashlib.sha256(compressed).hexdigest())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
