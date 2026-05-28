# 简单的生成器函数测试
def my_gen():
    yield 1
    yield 2
    yield 3

gen = my_gen()
print("Generator created:", gen)
val = gen.__next__()
print("First value:", val)