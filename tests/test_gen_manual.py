# 手动创建类似生成器表达式的函数
def make_gen(lst):
    for x in lst:
        yield x

gen = make_gen([1, 2, 3])
print("Generator created:", gen)
val = gen.__next__()
print("First value:", val)