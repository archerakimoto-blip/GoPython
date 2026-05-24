# 综合测试：try/except, with, yield

print("=" * 50)
print("Testing try/except")
print("=" * 50)

try:
    result = 1 / 0
except Exception as e:
    print("Caught error in try/except: ")
    print(e)

print("")
print("=" * 50)
print("Testing try/except/finally")
print("=" * 50)

try:
    x = 10 / 0
except Exception as e:
    print("In except block")
    print(e)
finally:
    print("In finally block (always runs)")

print("")
print("=" * 50)
print("Testing with statement")
print("=" * 50)

with open("test.txt", "r") as f:
    print("File opened: ")
    print(f)

print("")
print("=" * 50)
print("Testing yield (generator)")
print("=" * 50)

def simple_generator(n):
    for i in range(n):
        yield i

gen = simple_generator(5)
print("Generator created")

try:
    v1 = next(gen)
    print("First value: ")
    print(v1)
    v2 = next(gen)
    print("Second value: ")
    print(v2)
except Exception as e:
    print("Generator error: ")
    print(e)

print("")
print("=" * 50)
print("All tests completed!")
print("=" * 50)
