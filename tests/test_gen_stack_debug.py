# 测试生成器栈布局
def make_gen(x):
    y = x + 1
    print("x =", x)
    print("y =", y)
    yield x
    yield y

gen = make_gen(10)
print("Generator created")
val1 = gen.__next__()
print("Yielded:", val1)