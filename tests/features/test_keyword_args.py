
def greet(name, greeting="Hello", **kwargs):
    let msg = greeting + ", " + name
    for key, value in kwargs.items():
        msg = msg + "; " + key + ": " + str(value)
    return msg

let result1 = greet("Alice")
print(result1)  # Should print "Hello, Alice"

let result2 = greet("Bob", greeting="Hi")
print(result2)  # Should print "Hi, Bob"

let result3 = greet("Charlie", greeting="Hey", age=30, city="New York")
print(result3)  # Should print "Hey, Charlie; age: 30; city: New York"


# 测试字典推导式
let numbers = [1, 2, 3, 4, 5]
let squares = {x: x*x for x in numbers}
print(squares)  # Should print {1:1, 2:4, 3:9, 4:16, 5:25}

let even_squares = {x: x*x for x in numbers if x % 2 == 0}
print(even_squares)  # Should print {2:4, 4:16}
