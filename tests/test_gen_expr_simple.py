
# 测试简单的生成器函数
def make_gen():
    yield 2
    yield 4
    yield 6

gen = make_gen()
print("Gen created")
val1 = gen.__next__()
print("First:", val1)
val2 = gen.__next__()
print("Second:", val2)
