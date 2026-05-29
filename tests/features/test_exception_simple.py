# Simple Exception Chaining Test

print("Test 1: Basic exception chaining")
try:
    try:
        raise ValueError("original")
    except ValueError as e:
        raise TypeError("new") from e
except TypeError as te:
    print(f"Caught: {te}")
    print(f"Has cause: {te.__cause__ is not None}")
    if te.__cause__ is not None:
        print(f"Cause: {te.__cause__}")
    print("Test 1 passed!")

print("Test completed!")
