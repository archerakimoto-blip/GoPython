# Test closure
add5 = (lambda x: lambda y: x + y)(5)
result = add5(10)
print(result)
