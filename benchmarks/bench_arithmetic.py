# 算术运算性能测试
# 测试整数和浮点数的基本运算速度

import time

N = 1000000

# 整数加法
start = time.time()
x = 0
i = 0
while i < N:
    x = x + 1
    i = i + 1
elapsed_add = time.time() - start

# 整数乘法
start = time.time()
x = 1
i = 0
while i < N:
    x = x * 2
    i = i + 1
elapsed_mul = time.time() - start

# 整数除法
start = time.time()
x = 0
i = 0
while i < N:
    x = 1000000007 // 3
    i = i + 1
elapsed_div = time.time() - start

# 整数取模
start = time.time()
x = 0
i = 0
while i < N:
    x = 1000000007 % 3
    i = i + 1
elapsed_mod = time.time() - start

# 浮点数加法
start = time.time()
x = 0.0
i = 0
while i < N:
    x = x + 1.5
    i = i + 1
elapsed_fadd = time.time() - start

# 浮点数乘法
start = time.time()
x = 1.0
i = 0
while i < N:
    x = x * 1.000001
    i = i + 1
elapsed_fmul = time.time() - start

# 幂运算
start = time.time()
x = 0
i = 0
while i < 100000:
    x = 2 ** 10
    i = i + 1
elapsed_pow = time.time() - start

print("BENCHMARK_RESULT|arithmetic_int_add|{:.6f}".format(elapsed_add))
print("BENCHMARK_RESULT|arithmetic_int_mul|{:.6f}".format(elapsed_mul))
print("BENCHMARK_RESULT|arithmetic_int_div|{:.6f}".format(elapsed_div))
print("BENCHMARK_RESULT|arithmetic_int_mod|{:.6f}".format(elapsed_mod))
print("BENCHMARK_RESULT|arithmetic_float_add|{:.6f}".format(elapsed_fadd))
print("BENCHMARK_RESULT|arithmetic_float_mul|{:.6f}".format(elapsed_fmul))
print("BENCHMARK_RESULT|arithmetic_pow|{:.6f}".format(elapsed_pow))
