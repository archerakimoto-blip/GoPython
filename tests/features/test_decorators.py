
# 简单的装饰器测试
def log_decorator(func):
    def wrapper(*args, **kwargs):
        let result = func(*args, **kwargs)
        return result
    return wrapper

@log_decorator
def add(a, b):
    return a + b

let res1 = add(2, 3)
print(res1)

# 带参数的装饰器测试
def repeat(times):
    def decorator(func):
        def wrapper(*args, **kwargs):
            let result = None
            let i = 0
            while i < times:
                result = func(*args, **kwargs)
                i = i + 1
            return result
        return wrapper
    return decorator

@repeat(3)
def greet(name):
    print("Hello, " + name)
    return "Hi, " + name

let res2 = greet("Alice")
print(res2)
