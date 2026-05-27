print("=== 字典操作基准测试 ===")

print()
print("1. Create dictionary (1000 entries):")

d = {}
i = 0
while i < 1000:
    key = str(i)
    d[key] = i * 2
    i = i + 1
print(len(d))

print()
print("2. Dictionary lookup (1000 lookups):")

d = {}
i = 0
while i < 1000:
    key = str(i)
    d[key] = i * 2
    i = i + 1

total = 0
idx = 0
while idx < 1000:
    key = str(idx)
    total = total + d[key]
    idx = idx + 1
print(total)

print()
print("=== 字典操作测试完成 ===")
