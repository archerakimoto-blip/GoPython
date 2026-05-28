# 测试生成器函数是否能正确访问参数
def make_gen(lst):
    print("List length:", len(lst))
    for x in lst:
        yield x

gen = make_gen([1, 2, 3])
print("Generator created:", gen)
val = gen.__next__()
print("First value:", val)