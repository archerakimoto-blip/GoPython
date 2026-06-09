# 循环性能测试
# 测试 for/while 循环、列表推导式等性能

import time

N = 100000

# while 循环
start = time.time()
i = 0
x = 0
while i < N:
    x = x + 1
    i = i + 1
elapsed_while = time.time() - start

# for 循环遍历 range
start = time.time()
x = 0
for i in range(N):
    x = x + 1
elapsed_for_range = time.time() - start

# for 循环遍历列表
data = []
i = 0
while i < N:
    data.append(i)
    i = i + 1

start = time.time()
x = 0
for v in data:
    x = x + v
elapsed_for_list = time.time() - start

# 嵌套循环
start = time.time()
x = 0
i = 0
while i < 300:
    j = 0
    while j < 300:
        x = x + 1
        j = j + 1
    i = i + 1
elapsed_nested = time.time() - start

# 列表推导式
start = time.time()
result = [i * 2 for i in range(N)]
elapsed_listcomp = time.time() - start

# 带条件的列表推导式
start = time.time()
result = [i for i in range(N) if i % 2 == 0]
elapsed_listcomp_filter = time.time() - start

# break/continue 循环
start = time.time()
x = 0
i = 0
while i < N:
    if i % 3 == 0:
        i = i + 1
        continue
    x = x + 1
    i = i + 1
elapsed_continue = time.time() - start

print("BENCHMARK_RESULT|loop_while|{:.6f}".format(elapsed_while))
print("BENCHMARK_RESULT|loop_for_range|{:.6f}".format(elapsed_for_range))
print("BENCHMARK_RESULT|loop_for_list|{:.6f}".format(elapsed_for_list))
print("BENCHMARK_RESULT|loop_nested_300x300|{:.6f}".format(elapsed_nested))
print("BENCHMARK_RESULT|loop_listcomp|{:.6f}".format(elapsed_listcomp))
print("BENCHMARK_RESULT|loop_listcomp_filter|{:.6f}".format(elapsed_listcomp_filter))
print("BENCHMARK_RESULT|loop_continue|{:.6f}".format(elapsed_continue))
