#!/usr/bin/env python3
"""
GoPy Performance Benchmark Script
"""

import subprocess
import time
import os
import re

def run_gopy(file_path):
    """Run GoPy and return execution time in ms"""
    start = time.time()
    result = subprocess.run(
        ['go', 'run', 'cmd/gopy/main.go', file_path],
        capture_output=True,
        text=True,
        timeout=60
    )
    elapsed = time.time() - start

    # Extract instruction count from output
    instructions = None
    for line in result.stdout.split('\n'):
        match = re.search(r'Executed (\d+) instructions', line)
        if match:
            instructions = int(match.group(1))
            break

    return elapsed * 1000, instructions

def run_cpython(file_path):
    """Run CPython and return execution time in ms"""
    start = time.time()
    subprocess.run(
        ['python3', file_path],
        capture_output=True,
        timeout=60
    )
    elapsed = time.time() - start
    return elapsed * 1000

def main():
    tests = [
        ('tests/benchmarks/000_test.py', 'Simple Calculation'),
        ('tests/benchmarks/000_function_test.py', 'Function Calls'),
        ('tests/benchmarks/002_loop_simple.py', 'Loop Operations'),
    ]

    print("=" * 80)
    print("                     GoPy Performance Benchmark")
    print("=" * 80)
    print()
    print(f"Test Time: {time.strftime('%Y-%m-%d %H:%M:%S')}")
    print()

    print("| Test Case            | GoPy (ms) | Instructions | CPython (ms) | Ratio (GoPy/CPython) |")
    print("|---------------------|-----------|--------------|--------------|----------------------|")

    total_gopy = 0
    total_cpython = 0
    count = 0

    for file_path, name in tests:
        if not os.path.exists(file_path):
            print(f"| {name:19} | N/A       | N/A          | N/A          | N/A                  |")
            continue

        try:
            gopy_ms, instructions = run_gopy(file_path)
            cpython_ms = run_cpython(file_path)

            ratio = gopy_ms / cpython_ms if cpython_ms > 0 else float('inf')

            instr_str = str(instructions) if instructions else "N/A"
            ratio_str = f"{ratio:.2f}x"

            print(f"| {name:19} | {gopy_ms:9.1f} | {instr_str:12} | {cpython_ms:12.1f} | {ratio_str:20} |")

            total_gopy += gopy_ms
            total_cpython += cpython_ms
            count += 1

        except Exception as e:
            print(f"| {name:19} | Error: {str(e)[:40]:40} |")

    if count > 0:
        avg_gopy = total_gopy / count
        avg_cpython = total_cpython / count
        avg_ratio = avg_gopy / avg_cpython if avg_cpython > 0 else float('inf')

        print(f"|---------------------|-----------|--------------|--------------|----------------------|")
        print(f"| **Average**         | **{avg_gopy:7.1f}** | -            | **{avg_cpython:10.1f}** | **{avg_ratio:.2f}x**              |")

    print()
    print("=" * 80)
    print("Notes:")
    print("- GoPy execution time includes compilation + interpretation")
    print("- CPython uses optimized bytecode compilation (pysource compile)")
    print("- For production use, pre-compile GoPy bytecode for better performance")
    print("=" * 80)

if __name__ == '__main__':
    main()
