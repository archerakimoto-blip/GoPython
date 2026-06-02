class MyClass:
    @classmethod
    def create(cls, x):
        return cls(x)

    @staticmethod
    def add(a, b):
        return a + b

obj = MyClass(0)

class Base:
    @classmethod
    def who(cls):
        return cls.__name__

    @staticmethod
    def greet():
        return "hello"

class Child(MyClass, Base):
    pass

print(Child.who())
print(Child.greet())

c = Child.create(5)
print(c.__class__.__name__)
print("test_classmethod_staticmethod: PASS")
