# 调试生成器表达式
def create_gen():
    lst = [1, 2, 3]
    print("List created:", lst)
    gen = (x for x in lst)
    print("Generator created:", gen)
    return gen

gen = create_gen()
print("Gen type:", type(gen))
val = gen.__next__()
print("First value:", val)