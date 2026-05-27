#!/usr/bin/env python3
"""
循环操作基准测试
"""

def nested_loops():
    total = 0
    for i in range(100):
        for j in range(100):
            total += i + j
    return total

def list_traversal():
    lst = list(range(1000))
    total = 0
    for num in lst:
        total += num
    return total

def main():
    print("=== 循环操作基准测试 ===")
    
    # 嵌套循环测试
    print("\n1. Nested loops (100x100):")
    result = nested_loops()
    print(f"   Result: {result}")
    
    # 列表遍历测试
    print("\n2. List traversal (1000 elements):")
    result = list_traversal()
    print(f"   Result: {result}")
    
    # 简单循环测试
    print("\n3. Simple loop (100000 iterations):")
    count = 0
    for i in range(100000):
        count += 1
    print(f"   Count: {count}")
    
    print("\n=== 循环操作测试完成 ===")

if __name__ == "__main__":
    main()
