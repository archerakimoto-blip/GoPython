
def make_gen():
    yield 2

gen = make_gen()
x = gen.__next__()
