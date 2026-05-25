print("=== Lambda Expression Tests ===")

add = lambda x, y: x + y
print("Lambda add(5, 3):")
print(add(5, 3))

square = lambda x: x * x
print("Lambda square(5):")
print(square(5))

mul = lambda a, b: a * b
print("Lambda mul(4, 7):")
print(mul(4, 7))

add_ten = lambda x: x + 10
print("Lambda add_ten(5):")
print(add_ten(5))

print("\n=== Math Module Tests ===")
print("math.pi:")
print(math.pi)

print("math.e:")
print(math.e)

print("math.sqrt(16):")
print(math.sqrt(16))

print("math.sin(0):")
print(math.sin(0))

print("math.cos(0):")
print(math.cos(0))

print("math.floor(3.7):")
print(math.floor(3.7))

print("math.ceil(3.2):")
print(math.ceil(3.2))

print("math.abs(-5):")
print(math.abs(-5))

print("math.pow(2, 3):")
print(math.pow(2, 3))

print("\n=== String Formatting Tests ===")
print("format with %s:")
result = format("Hello, %s!", "World")
print(result)

print("format with %d:")
result = format("Value: %d", 42)
print(result)

print("format with %f:")
result = format("Pi: %f", 3.14159)
print(result)

print("format with multiple placeholders:")
result = format("%s has %d apples and %f oranges", "John", 5, 2.5)
print(result)

print("\n=== Built-in Functions Tests ===")
print("round(3.7):")
print(round(3.7))

print("round(3.14159, 2):")
print(round(3.14159, 2))

print("len of list:")
print(len([1, 2, 3, 4, 5]))

print("len of string:")
print(len("hello"))

print("len of dict:")
print(len({"a": 1, "b": 2}))

print("\nAll tests passed!")
