# Simple Exception Chaining Test

print("Test 1: Basic exception chaining")
try:
    try:
        raise ValueError("original")
    except ValueError as e:
        raise TypeError("new") from e
except TypeError as te:
    print("Caught: " + str(te))
    if te.__cause__ is not None:
        print("Has cause: True")
        print("Cause: " + str(te.__cause__))
    else:
        print("Has cause: False")
    print("Test 1 passed!")

print("Test completed!")
