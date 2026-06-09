# 函数调用性能测试
# 测试函数定义、调用、递归、闭包等性能

import time

N = 100000

# 简单函数调用
def simple_func():
    return 1

start = time.time()
i = 0
x = 0
while i < N:
    x = simple_func()
    i = i + 1
elapsed_simple = time.time() - start

# 带参数函数调用
def add_func(a, b):
    return a + b

start = time.time()
i = 0
x = 0
while i < N:
    x = add_func(i, i + 1)
    i = i + 1
elapsed_args = time.time() - start

# 递归斐波那契
def fib(n):
    if n <= 1:
        return n
    return fib(n - 1) + fib(n - 2)

start = time.time()
x = fib(25)
elapsed_fib = time.time() - start

# 闭包调用
def make_counter(start_val):
    count = start_val
    def increment():
        return count + 1
    return increment

counter = make_counter(0)
start = time.time()
i = 0
x = 0
while i < N:
    x = counter()
    i = i + 1
elapsed_closure = time.time() - start

# 嵌套函数调用
def inner():
    return 42

def outer():
    return inner()

start = time.time()
i = 0
x = 0
while i < N:
    x = outer()
    i = i + 1
elapsed_nested = time.time() - start

# varargs 函数调用
def varargs_func(*args):
    return len(args)

start = time.time()
i = 0
x = 0
while i < N:
    x = varargs_func(1, 2, 3, 4, 5)
    i = i + 1
elapsed_varargs = time.time() - start

print("BENCHMARK_RESULT|func_simple_call|{:.6f}".format(elapsed_simple))
print("BENCHMARK_RESULT|func_args_call|{:.6f}".format(elapsed_args))
print("BENCHMARK_RESULT|func_recursive_fib25|{:.6f}".format(elapsed_fib))
print("BENCHMARK_RESULT|func_closure_call|{:.6f}".format(elapsed_closure))
print("BENCHMARK_RESULT|func_nested_call|{:.6f}".format(elapsed_nested))
print("BENCHMARK_RESULT|func_varargs_call|{:.6f}".format(elapsed_varargs))
