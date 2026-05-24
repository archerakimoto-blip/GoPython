# 简单的生成器测试
print("=== 测试简单生成器 ===")

def simple_gen():
    yield 1
    yield 2
    yield 3

gen = simple_gen()
print("生成器已创建:", gen)

try:
    val1 = next(gen)
    print("第一次 next:", val1)
except Exception as e:
    print("第一次异常:", e)

try:
    val2 = next(gen)
    print("第二次 next:", val2)
except Exception as e:
    print("第二次异常:", e)

try:
    val3 = next(gen)
    print("第三次 next:", val3)
except Exception as e:
    print("第三次异常:", e)

try:
    val4 = next(gen)
    print("第四次 next:", val4)
except Exception as e:
    print("第四次异常:", e)

print("\n=== 测试完成 ===")
