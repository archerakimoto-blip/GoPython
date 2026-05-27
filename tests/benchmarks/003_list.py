#!/usr/bin/env python3
"""
列表操作基准测试
"""

def create_and_append():
    lst = []
    for i in range(10000):
        lst.append(i)
    return len(lst)

def list_comprehension():
    lst = [x * 2 for x in range(5000)]
    return len(lst)

def list_indexing():
    lst = list(range(10000))
    total = 0
    for i in range(1000):
        total += lst[i]
    return total

def main():
    print("=== 列表操作基准测试 ===")
    
    # 列表创建和追加测试
    print("\n1. Create and append (10000 elements):")
    result = create_and_append()
    print(f"   Length: {result}")
    
    # 列表推导式测试
    print("\n2. List comprehension (5000 elements):")
    result = list_comprehension()
    print(f"   Length: {result}")
    
    # 列表索引测试
    print("\n3. List indexing (1000 accesses):")
    result = list_indexing()
    print(f"   Total: {result}")
    
    print("\n=== 列表操作测试完成 ===")

if __name__ == "__main__":
    main()
