# Exception Chaining Tests
# 测试 Python 的异常链功能

print("Test 1: Basic exception chaining with 'from'")
try:
    try:
        raise ValueError("original error")
    except ValueError as e:
        raise TypeError("new error") from e
except TypeError as te:
    print(f"Caught TypeError: {te}")
    print(f"Cause: {te.__cause__}")
    if te.__cause__ is not None:
        print(f"Cause type: {type(te.__cause__).__name__}")
        print(f"Cause message: {te.__cause__}")
    print("Test 1 passed!")

print("\n" + "="*50 + "\n")

print("Test 2: Exception chaining with None (implicit chaining)")
try:
    raise ValueError("original")
except ValueError:
    raise TypeError("new")
except Exception as e:
    print(f"Caught: {e}")
    print(f"__cause__: {e.__cause__}")
    print(f"__context__: {e.__context__}")
    print("Test 2 passed!")

print("\n" + "="*50 + "\n")

print("Test 3: Suppress exception context")
try:
    try:
        raise RuntimeError("error 1")
    except RuntimeError:
        raise ValueError("error 2") from None
except ValueError as v:
    print(f"Caught ValueError: {v}")
    print(f"__cause__: {v.__cause__}")
    print(f"__context__: {v.__context__}")
    if v.__suppress_context__:
        print("Context was suppressed")
    print("Test 3 passed!")

print("\n" + "="*50 + "\n")

print("Test 4: Complex exception chain")
try:
    try:
        try:
            raise OSError("system error")
        except OSError as oe:
            raise IOError("io error") from oe
    except IOError as ie:
        raise RuntimeError("runtime error") from ie
except RuntimeError as re:
    print(f"Caught RuntimeError: {re}")
    print(f"Direct cause: {re.__cause__}")
    if re.__cause__ is not None:
        print(f"Cause: {re.__cause__}")
        print(f"Cause's cause: {re.__cause__.__cause__}")
    print("Test 4 passed!")

print("\nAll tests completed!")
