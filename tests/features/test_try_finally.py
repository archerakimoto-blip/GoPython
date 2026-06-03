# Test 1: try-only-finally with exception propagation
print("Test 1:")
try:
    try:
        print("inner try")
        raise "error1"
    finally:
        print("inner finally")
except:
    print("outer caught")

# Test 2: try-except-finally (should still work)
print("Test 2:")
try:
    print("try2")
    raise "error2"
except:
    print("except2")
finally:
    print("finally2")

# Test 3: try-only-finally without exception
print("Test 3:")
try:
    print("try3")
finally:
    print("finally3")

# Test 4: nested try-finally with exception in function
print("Test 4:")
def foo():
    try:
        print("foo try")
        raise "error4"
    finally:
        print("foo finally")

try:
    foo()
except:
    print("outer caught4")

# Test 5: try-except-finally with exception in function
print("Test 5:")
def bar():
    try:
        print("bar try")
        raise "error5"
    except:
        print("bar except")
    finally:
        print("bar finally")

bar()

# Test 6: multiple finally blocks
print("Test 6:")
try:
    try:
        print("inner6")
        raise "error6"
    finally:
        print("inner finally6")
except:
    print("outer caught6")
finally:
    print("outer finally6")

print("all tests done")
