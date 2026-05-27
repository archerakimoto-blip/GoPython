#!/usr/bin/env python3
"""
函数调用基准测试
"""

def simple_add(a, b):
    return a + b

def multiply(x, y, z):
    return x * y * z

def recursive_function(n):
    if n <= 0:
        return 1
    return recursive_function(n-1) + n

def function_call_benchmark():
    result = 0
    for i in range(10000):
        result = simple_add(i, i+1)
    return result

def main():
    print("=== 函数调用基准测试 ===")
    
    # 简单函数调用测试
    print("\n1. Simple function calls (10000):")
    result = function_call_benchmark()
    print(f"   Result: {result}")
    
    # 多参数函数调用
    print("\n2. Multiple parameters function:")
    result = multiply(2, 3, 4)
    print(f"   Result: {result}")
    
    # 递归函数调用
    print("\n3. Recursive function (fibonacci-like, n=10):")
    result = recursive_function(10)
    print(f"   Result: {result}")
    
    print("\n=== 函数调用测试完成 ===")

if __name__ == "__main__":
    main()
