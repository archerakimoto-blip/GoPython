# 数学运算性能测试
# 测试 math 模块和复杂计算的性能

import time
import math

N = 100000

# math.sqrt
start = time.time()
i = 0
x = 0.0
while i < N:
    x = math.sqrt(i + 1)
    i = i + 1
elapsed_sqrt = time.time() - start

# math.sin
start = time.time()
i = 0
x = 0.0
while i < N:
    x = math.sin(i * 0.001)
    i = i + 1
elapsed_sin = time.time() - start

# math.cos
start = time.time()
i = 0
x = 0.0
while i < N:
    x = math.cos(i * 0.001)
    i = i + 1
elapsed_cos = time.time() - start

# math.log
start = time.time()
i = 0
x = 0.0
while i < N:
    x = math.log(i + 1)
    i = i + 1
elapsed_log = time.time() - start

# math.pow
start = time.time()
i = 0
x = 0.0
while i < N:
    x = math.pow(i + 1, 0.5)
    i = i + 1
elapsed_pow = time.time() - start

# math.floor / math.ceil
start = time.time()
i = 0
x = 0
while i < N:
    x = math.floor(i * 1.5)
    i = i + 1
elapsed_floor = time.time() - start

# 素数筛 (Eratosthenes)
def sieve(n):
    flags = [True] * (n + 1)
    flags[0] = False
    flags[1] = False
    i = 2
    while i * i <= n:
        if flags[i]:
            j = i * i
            while j <= n:
                flags[j] = False
                j = j + i
        i = i + 1
    count = 0
    i = 2
    while i <= n:
        if flags[i]:
            count = count + 1
        i = i + 1
    return count

start = time.time()
x = sieve(100000)
elapsed_sieve = time.time() - start

print("BENCHMARK_RESULT|math_sqrt|{:.6f}".format(elapsed_sqrt))
print("BENCHMARK_RESULT|math_sin|{:.6f}".format(elapsed_sin))
print("BENCHMARK_RESULT|math_cos|{:.6f}".format(elapsed_cos))
print("BENCHMARK_RESULT|math_log|{:.6f}".format(elapsed_log))
print("BENCHMARK_RESULT|math_pow|{:.6f}".format(elapsed_pow))
print("BENCHMARK_RESULT|math_floor|{:.6f}".format(elapsed_floor))
print("BENCHMARK_RESULT|math_sieve_100k|{:.6f}".format(elapsed_sieve))
