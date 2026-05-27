#!/usr/bin/env python3
"""
字符串操作基准测试
"""

def string_concatenation():
    result = ""
    for i in range(1000):
        result += str(i)
    return len(result)

def string_formatting():
    result = ""
    for i in range(500):
        result += f"Number: {i}, Value: {i*2}"
    return len(result)

def string_comparison():
    count = 0
    str1 = "test"
    str2 = "test"
    for i in range(1000):
        if str1 == str2:
            count += 1
    return count

def main():
    print("=== 字符串操作基准测试 ===")
    
    # 字符串拼接测试
    print("\n1. String concatenation (1000 iterations):")
    result = string_concatenation()
    print(f"   Length: {result}")
    
    # 字符串格式化测试
    print("\n2. String formatting (500 iterations):")
    result = string_formatting()
    print(f"   Length: {result}")
    
    # 字符串比较测试
    print("\n3. String comparison (1000 iterations):")
    result = string_comparison()
    print(f"   Count: {result}")
    
    print("\n=== 字符串操作测试完成 ===")

if __name__ == "__main__":
    main()
