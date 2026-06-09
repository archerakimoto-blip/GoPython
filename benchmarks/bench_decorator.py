# 装饰器性能测试
# 测试装饰器的开销

import time

N = 50000

# 无装饰器基线
def bare_func(x):
    return x + 1

start = time.time()
i = 0
x = 0
while i < N:
    x = bare_func(i)
    i = i + 1
elapsed_bare = time.time() - start

# 简单装饰器
def simple_decorator(func):
    def wrapper(x):
        return func(x)
    return wrapper

@simple_decorator
def decorated_func(x):
    return x + 1

start = time.time()
i = 0
x = 0
while i < N:
    x = decorated_func(i)
    i = i + 1
elapsed_simple_deco = time.time() - start

# 多层装饰器
def deco1(func):
    def wrapper(x):
        return func(x)
    return wrapper

def deco2(func):
    def wrapper(x):
        return func(x)
    return wrapper

@deco1
@deco2
def multi_deco_func(x):
    return x + 1

start = time.time()
i = 0
x = 0
while i < N:
    x = multi_deco_func(i)
    i = i + 1
elapsed_multi_deco = time.time() - start

# property 装饰器
class PropClass:
    def __init__(self):
        self._value = 0

    @property
    def value(self):
        return self._value

start = time.time()
obj = PropClass()
i = 0
x = 0
while i < N:
    x = obj.value
    i = i + 1
elapsed_property = time.time() - start

print("BENCHMARK_RESULT|decorator_bare|{:.6f}".format(elapsed_bare))
print("BENCHMARK_RESULT|decorator_simple|{:.6f}".format(elapsed_simple_deco))
print("BENCHMARK_RESULT|decorator_multi|{:.6f}".format(elapsed_multi_deco))
print("BENCHMARK_RESULT|decorator_property|{:.6f}".format(elapsed_property))
