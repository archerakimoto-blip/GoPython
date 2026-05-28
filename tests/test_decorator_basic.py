def simple_decorator(fn):
    return fn

@simple_decorator
def test():
    pass

print("test completed")
