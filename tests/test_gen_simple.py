# 最简单的生成器测试
def gen():
    yield 1
    yield 2
g = gen()
print("g:", g)
print("type(g):", type(g))