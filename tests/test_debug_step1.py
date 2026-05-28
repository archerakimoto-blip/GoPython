
# 分步骤测试，每次只测试一个 __next__ 调用
def make_gen():
    yield 2
    yield 4

gen = make_gen()
print("Before first next")
x = gen.__next__()
print("Got first:", x)
