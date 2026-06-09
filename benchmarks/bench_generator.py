# 生成器性能测试
# 测试生成器创建和迭代性能

import time

N = 50000

# 简单生成器
def simple_gen(n):
    i = 0
    while i < n:
        yield i
        i = i + 1

start = time.time()
x = 0
for v in simple_gen(N):
    x = x + v
elapsed_simple_gen = time.time() - start

# 生成器表达式
start = time.time()
x = 0
for v in (i * 2 for i in range(N)):
    x = x + v
elapsed_gen_expr = time.time() - start

# yield from 委托
def inner_gen(n):
    i = 0
    while i < n:
        yield i
        i = i + 1

def outer_gen(n):
    yield from inner_gen(n)

start = time.time()
x = 0
for v in outer_gen(N):
    x = x + v
elapsed_yield_from = time.time() - start

# 生成器与列表对比
start = time.time()
x = 0
for v in [i for i in range(N)]:
    x = x + v
elapsed_list_vs_gen = time.time() - start

print("BENCHMARK_RESULT|generator_simple|{:.6f}".format(elapsed_simple_gen))
print("BENCHMARK_RESULT|generator_expression|{:.6f}".format(elapsed_gen_expr))
print("BENCHMARK_RESULT|generator_yield_from|{:.6f}".format(elapsed_yield_from))
print("BENCHMARK_RESULT|generator_list_compare|{:.6f}".format(elapsed_list_vs_gen))
