# GoPy 新功能综合测试

print("=" * 60)
print("测试异常处理 (try/except/finally)")
print("=" * 60)

def test_try_except():
    try:
        x = 1 / 0
        print("Should not reach here")
    except Exception as e:
        print("成功捕获异常:", e)
    finally:
        print("Finally 块执行")

test_try_except()

print("\n" + "=" * 60)
print("测试 with 语句")
print("=" * 60)

def test_with():
    with open("test.txt", "r") as f:
        print("文件已打开:", f)

test_with()

print("\n" + "=" * 60)
print("测试生成器 (yield)")
print("=" * 60)

def counter(n):
    for i in range(n):
        yield i

gen = counter(5)
print("生成器已创建")
print("第一个值:", next(gen))
print("第二个值:", next(gen))

print("\n" + "=" * 60)
print("测试新内置函数")
print("=" * 60)

print("type(42) =", type(42))
print("str(42) =", str(42))
print("int(3.7) =", int(3.7))
print("float(42) =", float(42))
print("bool(0) =", bool(0))
print("bool(1) =", bool(1))
print("abs(-10) =", abs(-10))
print("range(3) =", range(3))
print("min(1, 2, 3) =", min(1, 2, 3))
print("max(1, 2, 3) =", max(1, 2, 3))
print("sum([1, 2, 3, 4]) =", sum([1, 2, 3, 4]))

print("\n" + "=" * 60)
print("所有测试完成！")
print("=" * 60)
