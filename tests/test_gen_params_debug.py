# 调试参数传递
def make_gen(x):
    print("x =", x)
    print("x type:", type(x).__name__)
    yield x

# 测试不同类型的参数
print("=== Test with integer ===")
gen1 = make_gen(42)
val1 = gen1.__next__()
print("Yielded:", val1)

print("\n=== Test with string ===")
gen2 = make_gen("hello")
val2 = gen2.__next__()
print("Yielded:", val2)