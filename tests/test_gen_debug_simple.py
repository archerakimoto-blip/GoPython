
def make_gen():
    yield 2
    yield 4

gen = make_gen()
x = gen.__next__()
y = gen.__next__()
print("Done:", x, y)
