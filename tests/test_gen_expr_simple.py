# 最简单的生成器表达式测试
gen = (x for x in [1, 2, 3])
print("Generator created")
# 尝试获取生成器的第一个值
val = gen.__next__()
print("First value:", val)