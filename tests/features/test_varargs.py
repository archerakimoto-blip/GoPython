def greet(name, *args):
    print(name)
    if len(args) > 0:
        print(args)

greet("Alice")
greet("Bob", "Hello", "World")
greet("Charlie", 1, 2, 3)

def sum(*numbers):
    let total = 0
    for n in numbers:
        total = total + n
    return total

print(sum())
print(sum(1))
print(sum(1, 2, 3, 4, 5))
