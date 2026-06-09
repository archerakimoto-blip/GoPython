# 异常处理性能测试
# 测试 try/except 的开销

import time

N = 50000

# try/except 无异常
start = time.time()
i = 0
x = 0
while i < N:
    try:
        x = x + 1
    except:
        x = x - 1
    i = i + 1
elapsed_no_exception = time.time() - start

# try/except 有异常
start = time.time()
i = 0
x = 0
while i < N:
    try:
        x = x + 1
        if x < 0:
            raise ValueError("test")
    except:
        x = 0
    i = i + 1
elapsed_no_raise = time.time() - start

# 实际抛出和捕获异常
start = time.time()
i = 0
x = 0
while i < 10000:
    try:
        raise ValueError("test error")
    except ValueError:
        x = x + 1
    i = i + 1
elapsed_catch_exception = time.time() - start

# 嵌套 try/except
start = time.time()
i = 0
x = 0
while i < N:
    try:
        try:
            x = x + 1
        except:
            x = 0
    except:
        x = -1
    i = i + 1
elapsed_nested_try = time.time() - start

# try/finally
start = time.time()
i = 0
x = 0
while i < N:
    try:
        x = x + 1
    finally:
        x = x + 0
    i = i + 1
elapsed_try_finally = time.time() - start

print("BENCHMARK_RESULT|exception_no_raise|{:.6f}".format(elapsed_no_exception))
print("BENCHMARK_RESULT|exception_no_raise_path|{:.6f}".format(elapsed_no_raise))
print("BENCHMARK_RESULT|exception_catch|{:.6f}".format(elapsed_catch_exception))
print("BENCHMARK_RESULT|exception_nested_try|{:.6f}".format(elapsed_nested_try))
print("BENCHMARK_RESULT|exception_try_finally|{:.6f}".format(elapsed_try_finally))
