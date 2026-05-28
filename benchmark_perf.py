"""
Performance benchmark for comparison
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

import time

start = time.time()

i = 0
while i < 100:
    fibonacci(20)
    factorial(100)
    i = i + 1

end = time.time()
elapsed = (end - start) * 1000
print('Elapsed: %.2f ms' % elapsed)
