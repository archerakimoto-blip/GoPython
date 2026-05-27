
print("=== 列表操作基准测试 ===")
lst = []
i = 0
while i < 5000:
    lst.append(i)
    i = i + 1
print(len(lst))
print("=== 测试完成 ===")
