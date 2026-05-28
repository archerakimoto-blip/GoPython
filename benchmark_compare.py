"""
Performance comparison benchmark
"""

def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

def factorial(n):
    result = 1
    i = 1
    while i <= n:
        result = result * i
        i = i + 1
    return result

def loop_test(count):
    total = 0
    i = 0
    while i < count:
        total = total + i
        i = i + 1
    return total

def nested_loop(outer, inner):
    total = 0
    i = 0
    while i < outer:
        j = 0
        while j < inner:
            total = total + i * j
            j = j + 1
        i = i + 1
    return total

print("=== Performance Benchmark ===")
print()

# Test 1: Fibonacci
print("1. Fibonacci(25):")
result = fibonacci(25)
print("   Result:", result)

# Test 2: Factorial
print("2. Factorial(500):")
result = factorial(500)
print("   Result length:", len(str(result)))

# Test 3: Simple loop
print("3. Loop (100K iterations):")
result = loop_test(100000)
print("   Result:", result)

# Test 4: Nested loop
print("4. Nested loop (100x100):")
result = nested_loop(100, 100)
print("   Result:", result)

print()
print("=== Benchmark Complete ===")
