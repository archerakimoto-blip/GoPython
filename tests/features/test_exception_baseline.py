print("Test: Basic exception chaining")
try:
    raise ValueError("test")
except ValueError:
    print("Caught!")
print("Done!")
