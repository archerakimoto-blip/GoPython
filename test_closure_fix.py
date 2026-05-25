# Test nested closure
result = ((lambda x: lambda y: x + y)(5))(10)
print(result)
