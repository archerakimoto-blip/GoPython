
# 测试生成器表达式
gen = (x * 2 for x in [1, 2, 3])
x = gen.__next__()
