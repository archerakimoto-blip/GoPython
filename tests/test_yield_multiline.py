def counter(n):
    for i in range(n):
        yield i

gen = counter(3)
print("Generator created")

try:
    val1 = next(gen)
    print(val1)
except Exception as e:
    print("Error: ")
    print(e)
