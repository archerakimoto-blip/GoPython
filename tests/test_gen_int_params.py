# 测试带整数参数的生成器
def make_gen(x):
    print("x =", x)
    yield x

gen1 = make_gen(42)
val1 = gen1.__next__()
print("Yielded:", val1)