print("Test: Basic try-except")
try:
    raise Exception("test")
except Exception:
    print("Caught!")
print("Done!")
