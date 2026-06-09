# 变量与作用域性能测试
# 测试变量绑定、作用域查找、global/nonlocal 等性能

import time

N = 100000

# 局部变量访问
start = time.time()
x = 0
i = 0
while i < N:
    x = x + 1
    i = i + 1
elapsed_local = time.time() - start

# 嵌套作用域变量访问
def outer_scope():
    outer_var = 42
    def inner():
        return outer_var
    return inner

fn = outer_scope()
start = time.time()
i = 0
x = 0
while i < N:
    x = fn()
    i = i + 1
elapsed_scope = time.time() - start

# 多重赋值
start = time.time()
i = 0
while i < N:
    a, b, c = 1, 2, 3
    i = i + 1
elapsed_multi_assign = time.time() - start

# 增强赋值
start = time.time()
x = 0
i = 0
while i < N:
    x += 1
    i += 1
elapsed_aug_assign = time.time() - start

# global 变量访问
global_var = 100

def read_global():
    return global_var

start = time.time()
i = 0
x = 0
while i < N:
    x = read_global()
    i = i + 1
elapsed_global = time.time() - start

# del 语句
start = time.time()
i = 0
while i < N:
    d = {"key": "value"}
    del d["key"]
    i = i + 1
elapsed_del = time.time() - start

print("BENCHMARK_RESULT|scope_local|{:.6f}".format(elapsed_local))
print("BENCHMARK_RESULT|scope_nested|{:.6f}".format(elapsed_scope))
print("BENCHMARK_RESULT|scope_multi_assign|{:.6f}".format(elapsed_multi_assign))
print("BENCHMARK_RESULT|scope_aug_assign|{:.6f}".format(elapsed_aug_assign))
print("BENCHMARK_RESULT|scope_global|{:.6f}".format(elapsed_global))
print("BENCHMARK_RESULT|scope_del|{:.6f}".format(elapsed_del))
