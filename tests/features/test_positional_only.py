# 测试仅位置参数 (/) — Python 3.8+

# 1. 基本仅位置参数
def greet(name, /, greeting="Hello"):
    return greeting + " " + name

print(greet("World"))
print(greet("World", "Hi"))

# 2. 仅位置参数不能作为关键字传递
try:
    greet(name="World")
    print("FAIL: should not allow keyword for positional-only param")
except:
    print("cannot pass positional-only as keyword")

# 3. / 后的参数可以作为关键字传递
print(greet("World", greeting="Hey"))

# 4. / 和 * 组合
def func(a, b, /, c, d=10, *, e, f=20):
    return str(a + b + c + d + e + f)

print(func(1, 2, 3, e=4))

# 5. 所有参数都是仅位置参数
def all_pos(x, y, /):
    return x * y

print(all_pos(3, 4))

try:
    all_pos(x=3, y=4)
    print("FAIL: should not allow keyword for all-positional-only")
except:
    print("cannot pass keyword for all-positional-only params")

# 6. 仅位置参数带默认值
def with_default(a, b=10, /):
    return a + b

print(with_default(1))
print(with_default(1, 2))

try:
    with_default(a=1)
    print("FAIL: should not allow keyword for positional-only with default")
except:
    print("cannot pass keyword for positional-only with default")

print("test_positional_only: PASS")
