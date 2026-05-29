def identity(x):
    return x

agen = identity([1, 2, 3, 4, 5])
result = [x async for x in agen if x > 2]
print("Async list comp result:", result)
print("Test passed:", len(result) == 3)
print("Result values:", result[0] == 3, result[1] == 4, result[2] == 5)
