def test_gen():
    gen = (x for x in [1, 2, 3])
    print("Generator created")
    # 直接调用生成器的 __next__ 方法
    print("First:", gen.__next__())
    print("Second:", gen.__next__())
    print("Third:", gen.__next__())

test_gen()