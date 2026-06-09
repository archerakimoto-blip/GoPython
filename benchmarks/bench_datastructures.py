# 数据结构性能测试
# 测试列表、字典、集合的创建、访问、修改性能

import time

N = 50000

# 列表创建
start = time.time()
lst = []
i = 0
while i < N:
    lst.append(i)
    i = i + 1
elapsed_list_create = time.time() - start

# 列表索引访问
start = time.time()
x = 0
i = 0
while i < N:
    x = lst[i]
    i = i + 1
elapsed_list_access = time.time() - start

# 列表索引赋值
start = time.time()
i = 0
while i < N:
    lst[i] = i * 2
    i = i + 1
elapsed_list_assign = time.time() - start

# 列表切片
start = time.time()
i = 0
while i < 1000:
    s = lst[100:500]
    i = i + 1
elapsed_list_slice = time.time() - start

# 字典创建
start = time.time()
d = {}
i = 0
while i < N:
    d[i] = i * 2
    i = i + 1
elapsed_dict_create = time.time() - start

# 字典查找
start = time.time()
x = 0
i = 0
while i < N:
    x = d[i]
    i = i + 1
elapsed_dict_lookup = time.time() - start

# 字典删除
start = time.time()
i = 0
while i < N:
    del d[i]
    i = i + 1
elapsed_dict_delete = time.time() - start

# 集合创建
start = time.time()
s = set()
i = 0
while i < N:
    s.add(i)
    i = i + 1
elapsed_set_create = time.time() - start

# 集合成员检测
start = time.time()
x = False
i = 0
while i < N:
    x = i in s
    i = i + 1
elapsed_set_member = time.time() - start

# 字符串拼接
start = time.time()
result = ""
i = 0
while i < 10000:
    result = result + "x"
    i = i + 1
elapsed_str_concat = time.time() - start

print("BENCHMARK_RESULT|ds_list_create|{:.6f}".format(elapsed_list_create))
print("BENCHMARK_RESULT|ds_list_access|{:.6f}".format(elapsed_list_access))
print("BENCHMARK_RESULT|ds_list_assign|{:.6f}".format(elapsed_list_assign))
print("BENCHMARK_RESULT|ds_list_slice|{:.6f}".format(elapsed_list_slice))
print("BENCHMARK_RESULT|ds_dict_create|{:.6f}".format(elapsed_dict_create))
print("BENCHMARK_RESULT|ds_dict_lookup|{:.6f}".format(elapsed_dict_lookup))
print("BENCHMARK_RESULT|ds_dict_delete|{:.6f}".format(elapsed_dict_delete))
print("BENCHMARK_RESULT|ds_set_create|{:.6f}".format(elapsed_set_create))
print("BENCHMARK_RESULT|ds_set_member|{:.6f}".format(elapsed_set_member))
print("BENCHMARK_RESULT|ds_str_concat|{:.6f}".format(elapsed_str_concat))
