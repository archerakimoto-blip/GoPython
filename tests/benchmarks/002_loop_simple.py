"""
Loop and Control Flow Benchmark (Simplified for GoPy)
Tests basic loop structures and control flow operations
"""

def test_while_loop(n):
    result = 0
    i = 0
    while i < n:
        result = result + i
        i = i + 1
    return result

def test_nested_loops(size):
    result = 0
    i = 0
    while i < size:
        j = 0
        while j < size:
            result = result + i * j
            j = j + 1
        i = i + 1
    return result

def test_list_iteration(n):
    lst = []
    i = 0
    while i < n:
        lst.append(i)
        i = i + 1
    total = 0
    for item in lst:
        total = total + item
    return total

# Run benchmarks
print("=== Loop and Control Flow Benchmark ===")

print("1. Simple while loop (100,000 times):")
result1 = test_while_loop(100000)
print("  Result:", result1)

print()
print("2. Nested loops (100x100):")
result2 = test_nested_loops(100)
print("  Result:", result2)

print()
print("3. List iteration (5,000 elements):")
result3 = test_list_iteration(5000)
print("  Result:", result3)

print()
print("=== Loop Test Complete ===")
