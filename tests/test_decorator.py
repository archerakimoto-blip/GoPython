# Test basic decorator
def my_decorator(func):
    def wrapper():
        print("Before function")
        func()
        print("After function")
    return wrapper

@my_decorator
def say_hello():
    print("Hello!")

say_hello()

# Test decorator with arguments
def with_args_decorator(prefix):
    def decorator(func):
        def wrapper(*args, **kwargs):
            print(prefix + " Calling function")
            result = func(*args, **kwargs)
            print(prefix + " Done")
            return result
        return wrapper
    return decorator

@with_args_decorator(">>>")
def greet(name):
    print("Hello, " + name + "!")

greet("World")

# Test multiple decorators (applied bottom to top)
def decorator1(func):
    def wrapper():
        print("decorator1")
        func()
    return wrapper

def decorator2(func):
    def wrapper():
        print("decorator2")
        func()
    return wrapper

@decorator1
@decorator2
def test_func():
    print("test_func")

test_func()
