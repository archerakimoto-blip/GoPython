# 测试生成器内部的参数访问
def make_gen(x):
    print("x is:", x)
    yield x
    yield x + 1

gen = make_gen(42)
print("Generator created:", gen)
val1 = gen.__next__()
print("First value:", val1)
val2 = gen.__next__()
print("Second value:", val2)