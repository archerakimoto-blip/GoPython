def logger(fn):
    print("Before calling " + fn)
    result = fn()
    print("After calling " + fn)
    return result

@logger
def greet():
    return "Hello!"

result = greet
print("Greeting function assigned")
