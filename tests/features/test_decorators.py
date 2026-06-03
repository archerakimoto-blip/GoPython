
# 测试1: 简单装饰器 + varargs/kwargs
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

# 测试2: varargs 直接调用
def varargs_func(*args):
    return args

let r = varargs_func(1, 2, 3)
print(r)

# 测试3: varargs + kwargs 直接调用
def varargs_kwargs_func(*args, **kwargs):
    return args

let r2 = varargs_kwargs_func(1, 2, 3)
print(r2)

# 测试4: 闭包 + varargs（无 kwargs）
def make_adder(n):
    def adder(*args):
        let result = 0
        let i = 0
        while i < len(args):
            result = result + args[i]
            i = i + 1
        return result + n
    return adder

let add5 = make_adder(5)
print(add5(1, 2, 3))
