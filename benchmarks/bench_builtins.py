# 内置函数性能测试
# 测试常用内置函数的性能

import time

N = 50000

# len()
data = [i for i in range(1000)]
start = time.time()
i = 0
x = 0
while i < N:
    x = len(data)
    i = i + 1
elapsed_len = time.time() - start

# range()
start = time.time()
i = 0
while i < N:
    r = range(100)
    i = i + 1
elapsed_range = time.time() - start

# int() 类型转换
start = time.time()
i = 0
x = 0
while i < N:
    x = int("42")
    i = i + 1
elapsed_int_conv = time.time() - start

# str() 类型转换
start = time.time()
i = 0
x = ""
while i < N:
    x = str(i)
    i = i + 1
elapsed_str_conv = time.time() - start

# abs()
start = time.time()
i = 0
x = 0
while i < N:
    x = abs(-i)
    i = i + 1
elapsed_abs = time.time() - start

# max()/min()
data = [i for i in range(100)]
start = time.time()
i = 0
x = 0
while i < N:
    x = max(data)
    i = i + 1
elapsed_max = time.time() - start

# sorted()
data = [100 - i for i in range(100)]
start = time.time()
i = 0
x = []
while i < 1000:
    x = sorted(data)
    i = i + 1
elapsed_sorted = time.time() - start

# enumerate
data = [i for i in range(100)]
start = time.time()
x = 0
for idx, val in enumerate(data):
    x = x + val
elapsed_enumerate = time.time() - start

# zip
a = [i for i in range(100)]
b = [i * 2 for i in range(100)]
start = time.time()
x = 0
for va, vb in zip(a, b):
    x = x + va + vb
elapsed_zip = time.time() - start

# map
start = time.time()
x = 0
for v in map(lambda x: x * 2, range(1000)):
    x = x + v
elapsed_map = time.time() - start

# filter
start = time.time()
x = 0
for v in filter(lambda x: x % 2 == 0, range(1000)):
    x = x + v
elapsed_filter = time.time() - start

print("BENCHMARK_RESULT|builtin_len|{:.6f}".format(elapsed_len))
print("BENCHMARK_RESULT|builtin_range|{:.6f}".format(elapsed_range))
print("BENCHMARK_RESULT|builtin_int_conv|{:.6f}".format(elapsed_int_conv))
print("BENCHMARK_RESULT|builtin_str_conv|{:.6f}".format(elapsed_str_conv))
print("BENCHMARK_RESULT|builtin_abs|{:.6f}".format(elapsed_abs))
print("BENCHMARK_RESULT|builtin_max|{:.6f}".format(elapsed_max))
print("BENCHMARK_RESULT|builtin_sorted|{:.6f}".format(elapsed_sorted))
print("BENCHMARK_RESULT|builtin_enumerate|{:.6f}".format(elapsed_enumerate))
print("BENCHMARK_RESULT|builtin_zip|{:.6f}".format(elapsed_zip))
print("BENCHMARK_RESULT|builtin_map|{:.6f}".format(elapsed_map))
print("BENCHMARK_RESULT|builtin_filter|{:.6f}".format(elapsed_filter))
