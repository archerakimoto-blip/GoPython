def dec(func):
    return func

@dec
def foo():
    return 1

print(foo())
