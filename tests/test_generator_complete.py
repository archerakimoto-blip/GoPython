# 测试生成器的基本功能
def test_basic_generator():
    print("=== 测试基本生成器 ===")
    
    def count_to(n):
        for i in range(1, n+1):
            yield i
    
    gen = count_to(3)
    print("生成器已创建")
    
    try:
        val1 = next(gen)
        print("第一次 next():", val1)
    except Exception as e:
        print("第一次 next() 异常:", e)
    
    try:
        val2 = next(gen)
        print("第二次 next():", val2)
    except Exception as e:
        print("第二次 next() 异常:", e)
    
    try:
        val3 = next(gen)
        print("第三次 next():", val3)
    except Exception as e:
        print("第三次 next() 异常:", e)
    
    try:
        val4 = next(gen)
        print("第四次 next():", val4)
    except Exception as e:
        print("第四次 next() 异常:", e)

test_basic_generator()

# 测试斐波那契数列生成器
def test_fibonacci():
    print("\n=== 测试斐波那契生成器 ===")
    
    def fibonacci(n):
        a, b = 0, 1
        for _ in range(n):
            yield b
            a, b = b, a + b
    
    fib = fibonacci(5)
    print("斐波那契生成器已创建")
    
    for i in range(6):
        try:
            val = next(fib)
            print(f"第 {i+1} 个值: {val}")
        except Exception as e:
            print(f"第 {i+1} 个值异常: {e}")
            break

test_fibonacci()

# 测试不返回任何值的生成器
def test_empty_generator():
    print("\n=== 测试空生成器 ===")
    
    def empty_gen():
        pass
    
    gen = empty_gen()
    try:
        next(gen)
    except Exception as e:
        print("next() 异常:", e)

test_empty_generator()

# 测试只有一个 yield 的生成器
def test_single_yield():
    print("\n=== 测试单个 yield ===")
    
    def single():
        yield "hello"
    
    gen = single()
    val1 = next(gen)
    print("第一次 next():", val1)
    try:
        val2 = next(gen)
        print("第二次 next():", val2)
    except Exception as e:
        print("第二次 next() 异常:", e)

test_single_yield()

print("\n=== 所有测试完成 ===")
