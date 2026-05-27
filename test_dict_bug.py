
d = {}
i = 0
while i < 5:
    key = str(i)
    d[key] = i * 2
    i = i + 1
print(len(d))  # should print 5
print(d)
print("---")

# Test lookup
total = 0
idx = 0
while idx < 5:
    key = str(idx)
    total += d[key]
    idx += 1
print("Total:", total)

# Test list
a = [1, 2, 3]
a[0] = 5
a[1] = 10
print(a)
