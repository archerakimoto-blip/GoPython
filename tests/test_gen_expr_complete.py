
# 测试生成器表达式完整迭代
gen = (x * 2 for x in [1, 2, 3])
print("Generator created")
val1 = gen.__next__()
print("First value:", val1)
val2 = gen.__next__()
print("Second value:", val2)
val3 = gen.__next__()
print("Third value:", val3)
