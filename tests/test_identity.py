a = [1, 2, 3]
b = [1, 2, 3]
c = a

print("a is a:", a is a)
print("a is b:", a is b)
print("a is c:", a is c)

print("a is not a:", a is not a)
print("a is not b:", a is not b)
print("a is not c:", a is not c)

x = None
y = None
print("None is None:", x is y)

s1 = "hello"
s2 = "hello"
print("s1 is s2:", s1 is s2)