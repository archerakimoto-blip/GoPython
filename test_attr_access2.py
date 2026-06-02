class MyClass:
    @classmethod
    def create(cls, x):
        return cls(x)

obj = MyClass(100)
print(obj.__class__.__name__)
