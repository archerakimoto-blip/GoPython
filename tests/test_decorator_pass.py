def my_decorator(func):
    return func

@my_decorator
def say():
    pass

print("done")
