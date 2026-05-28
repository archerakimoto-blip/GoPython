# Decorator without function call
def decorator1(func):
    return func

@decorator1
def say():
    pass

print("done")
