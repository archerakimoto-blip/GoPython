#!/usr/bin/env python3
"""
字典操作基准测试
"""

def create_dict():
    d = {}
    for i in range(1000):
        d[str(i)] = i * 2
    return len(d)

def dict_lookup():
    d = {}
    for i in range(1000):
        d[str(i)] = i * 2
    
    total = 0
    for i in range(1000):
        total += d[str(i)]
    return total

def dict_update():
    d = {}
    for i in range(1000):
        d[str(i)] = i
    
    # 更新字典中的值
    for i in range(1000):
        d[str(i)] = i * 3
    
    return len(d)

def main():
    print("=== 字典操作基准测试 ===")
    
    # 字典创建测试
    print("\n1. Create dictionary (1000 entries):")
    result = create_dict()
    print(f"   Entries: {result}")
    
    # 字典查找测试
    print("\n2. Dictionary lookup (1000 lookups):")
    result = dict_lookup()
    print(f"   Total: {result}")
    
    # 字典更新测试
    print("\n3. Dictionary update (1000 updates):")
    result = dict_update()
    print(f"   Entries: {result}")
    
    print("\n=== 字典操作测试完成 ===")

if __name__ == "__main__":
    main()
