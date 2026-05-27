
print("=== 字符串操作基准测试 ===")

print()
print("1. String concatenation (1000 iterations):")

result = ""
i = 0
while i < 1000:
    result = result + str(i)
    i = i + 1
print(len(result))

print()
print("2. String comparison (1000 iterations):")

count = 0
str1 = "test"
str2 = "test"
i = 0
while i < 1000:
    if str1 == str2:
        count = count + 1
    i = i + 1
print(count)

print()
print("=== 字符串操作测试完成 ===")
