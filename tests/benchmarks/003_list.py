print("=== 列表操作基准测试 ===")

print()
print("1. Create and append (10000 elements):")

lst = []
i = 0
while i < 10000:
    lst.append(i)
    i = i + 1
print(len(lst))

print()
print("2. List indexing (1000 accesses):")

lst = []
i = 0
while i < 10000:
    lst.append(i)
    i = i + 1

total = 0
idx = 0
while idx < 1000:
    total = total + lst[idx]
    idx = idx + 1
print(total)

print()
print("=== 列表操作测试完成 ===")
