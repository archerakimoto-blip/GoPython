# 字符串操作性能测试
# 测试字符串创建、拼接、格式化、切片等性能

import time

N = 50000

# 字符串创建
start = time.time()
i = 0
while i < N:
    s = "hello world"
    i = i + 1
elapsed_create = time.time() - start

# 字符串拼接
start = time.time()
result = ""
i = 0
while i < 10000:
    result = result + "a"
    i = i + 1
elapsed_concat = time.time() - start

# 字符串格式化 (%)
start = time.time()
i = 0
while i < N:
    s = "value: %d" % i
    i = i + 1
elapsed_format_percent = time.time() - start

# 字符串长度
start = time.time()
s = "a" * 1000
i = 0
x = 0
while i < N:
    x = len(s)
    i = i + 1
elapsed_len = time.time() - start

# 字符串切片
start = time.time()
s = "abcdefghij" * 100
i = 0
x = ""
while i < N:
    x = s[10:20]
    i = i + 1
elapsed_slice = time.time() - start

# 字符串比较
start = time.time()
s1 = "hello world this is a test string"
s2 = "hello world this is a test string"
i = 0
x = False
while i < N:
    x = s1 == s2
    i = i + 1
elapsed_compare = time.time() - start

# 字符串重复
start = time.time()
i = 0
x = ""
while i < N:
    x = "ab" * 100
    i = i + 1
elapsed_repeat = time.time() - start

print("BENCHMARK_RESULT|string_create|{:.6f}".format(elapsed_create))
print("BENCHMARK_RESULT|string_concat|{:.6f}".format(elapsed_concat))
print("BENCHMARK_RESULT|string_format_percent|{:.6f}".format(elapsed_format_percent))
print("BENCHMARK_RESULT|string_len|{:.6f}".format(elapsed_len))
print("BENCHMARK_RESULT|string_slice|{:.6f}".format(elapsed_slice))
print("BENCHMARK_RESULT|string_compare|{:.6f}".format(elapsed_compare))
print("BENCHMARK_RESULT|string_repeat|{:.6f}".format(elapsed_repeat))
