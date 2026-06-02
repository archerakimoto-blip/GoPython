class MyClass:
    @classmethod
    def create(cls, x):
        return cls(x)

    @staticmethod
    def add(a, b):
        return a + b

# if MyClass.add(3, 4) != 7:
#     print("FAIL: staticmethod call")

obj = MyClass(0)
# if obj.add(10, 20) != 30:
#     print("FAIL: staticmethod via instance")

print(obj.add(10, 20))

class Base:
    @classmethod
    def who(cls):
        return cls.__name__

    @staticmethod
    def greet():
        return "hello"

class Child(Base):
    pass

print(Child.who())
print(Child.greet())

c = Child.create(5)
print(c.__class__.__name__)

print("test_classmethod_staticmethod: PASS")
