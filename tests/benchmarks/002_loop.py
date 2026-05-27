print("=== 循环操作基准测试 ===")

print()
print("1. Nested loops (100x100):")

total = 0
i = 0
while i < 100:
    j = 0
    while j < 100:
        total = total + i + j
        j = j + 1
    i = i + 1
print(total)

print()
print("2. List traversal (1000 elements):")

lst = []
i = 0
while i < 1000:
    lst.append(i)
    i = i + 1

total = 0
idx = 0
while idx < len(lst):
    total = total + lst[idx]
    idx = idx + 1
print(total)

print()
print("3. Simple loop (100000 iterations):")

count = 0
i = 0
while i < 100000:
    count = count + 1
    i = i + 1
print(count)

print()
print("=== 循环操作测试完成 ===")
