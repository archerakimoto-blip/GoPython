#!/usr/bin/env python3
"""
GoPy Performance Benchmark Script
"""

import subprocess
import time
import os
import re

def run_gopy(file_path, use_fast=False, use_jit=False):
    """Run GoPy and return execution time in ms"""
    args = ['go', 'run', 'cmd/gopy/main.go']
    if use_fast:
        args.append('--fast')
    if use_jit:
        args.append('--jit')
    args.append(file_path)

    start = time.time()
    result = subprocess.run(args, capture_output=True, text=True, timeout=60)
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
    subprocess.run(['python3', file_path], capture_output=True, timeout=60)
    elapsed = time.time() - start
    return elapsed * 1000

def main():
    tests = [
        ('tests/benchmarks/000_test.py', 'Simple Calculation'),
        ('tests/benchmarks/000_function_test.py', 'Function Calls'),
        ('tests/benchmarks/002_loop_simple.py', 'Loop Operations'),
    ]

    print("=" * 120)
    print("                     GoPy Performance Benchmark (Complete Comparison)")
    print("=" * 120)
    print()
    print(f"Test Time: {time.strftime('%Y-%m-%d %H:%M:%S')}")
    print()

    print("| Test Case            | GoPy | GoPy+JIT | FastVM | FastVM+JIT | CPython | Ratio G/CP | Ratio F/CP | Ratio FJ/CP |")
    print("|---------------------|------|----------|--------|------------|---------|------------|------------|-------------|")

    total_gopy = 0
    total_gopy_jit = 0
    total_fastvm = 0
    total_fastvm_jit = 0
    total_cpython = 0
    count = 0

    for file_path, name in tests:
        if not os.path.exists(file_path):
            continue

        try:
            gopy_ms, instructions = run_gopy(file_path, use_fast=False, use_jit=False)
            gopy_jit_ms, _ = run_gopy(file_path, use_fast=False, use_jit=True)
            fastvm_ms, _ = run_gopy(file_path, use_fast=True, use_jit=False)
            fastvm_jit_ms, _ = run_gopy(file_path, use_fast=True, use_jit=True)
            cpython_ms = run_cpython(file_path)

            ratio_g = gopy_ms / cpython_ms if cpython_ms > 0 else 0
            ratio_f = fastvm_ms / cpython_ms if cpython_ms > 0 else 0
            ratio_fj = fastvm_jit_ms / cpython_ms if cpython_ms > 0 else 0

            total_gopy += gopy_ms
            total_gopy_jit += gopy_jit_ms
            total_fastvm += fastvm_ms
            total_fastvm_jit += fastvm_jit_ms
            total_cpython += cpython_ms
            count += 1

            instr_str = str(instructions) if instructions else "N/A"
            
            print(f"| {name:19} | {gopy_ms:5.1f}ms | {gopy_jit_ms:8.1f}ms | {fastvm_ms:6.1f}ms | {fastvm_jit_ms:10.1f}ms | {cpython_ms:7.1f}ms | {ratio_g:10.2f}x | {ratio_f:10.2f}x | {ratio_fj:11.2f}x |")

        except Exception as e:
            print(f"| {name:19} | Error: {str(e)[:30]:30} |")

    if count > 0:
        avg_gopy = total_gopy / count
        avg_gopy_jit = total_gopy_jit / count
        avg_fastvm = total_fastvm / count
        avg_fastvm_jit = total_fastvm_jit / count
        avg_cpython = total_cpython / count
        avg_ratio_g = avg_gopy / avg_cpython if avg_cpython > 0 else 0
        avg_ratio_f = avg_fastvm / avg_cpython if avg_cpython > 0 else 0
        avg_ratio_fj = avg_fastvm_jit / avg_cpython if avg_cpython > 0 else 0

        print("|---------------------|------|----------|--------|------------|---------|------------|------------|-------------|")
        print(f"| **Average**         | **{avg_gopy:4.1f}ms** | **{avg_gopy_jit:7.1f}ms** | **{avg_fastvm:5.1f}ms** | **{avg_fastvm_jit:9.1f}ms** | **{avg_cpython:6.1f}ms** | **{avg_ratio_g:9.2f}x** | **{avg_ratio_f:9.2f}x** | **{avg_ratio_fj:10.2f}x** |")

    print()
    print("=" * 120)
    print("Notes:")
    print("- GoPy: Standard VM with debugging removed")
    print("- GoPy+JIT: Standard VM with JIT enabled")
    print("- FastVM: Optimized VM interpreter")
    print("- FastVM+JIT: Optimized VM with JIT enabled")
    print("- GoPy execution time includes compilation + interpretation")
    print("- CPython uses optimized bytecode compilation")
    print("- Ratio < 1 means faster than CPython")
    print("=" * 120)

if __name__ == '__main__':
    main()
