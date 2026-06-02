class MyClass:
    @classmethod
    def create(cls, x):
        return cls(x)

    @staticmethod
    def add(a, b):
        return a + b

print(MyClass.add(3, 4))
obj = MyClass(0)
print(obj.add(10, 20))
