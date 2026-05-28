def add_logging(fn):
    print("Logging enabled for " + fn)
    return fn

def add_counting(fn):
    count = 0
    def wrapper():
        nonlocal count
        count = count + 1
        print("Called " + str(count) + " times")
        return fn()
    return wrapper

@add_logging
def greet():
    print("Hello!")

@add_counting
@add_logging
def farewell():
    print("Goodbye!")

greet()
farewell()
farewell()
