class MyClass:
    @classmethod
    def create(cls, x):
        return cls(x)

    @staticmethod
    def add(a, b):
        return a + b

if MyClass.add(3, 4) != 7:
    print("FAIL: staticmethod call")

obj = MyClass(0)
if obj.add(10, 20) != 30:
    print("FAIL: staticmethod via instance")

class Base:
    @classmethod
    def who(cls):
        return cls.__name__

    @staticmethod
    def greet():
        return "hello"

class Child(MyClass, Base):
    pass

if Child.who() != "Child":
    print("FAIL: classmethod inheritance got " + str(Child.who()))
if Child.greet() != "hello":
    print("FAIL: staticmethod inheritance")

c = Child.create(5)
if c.__class__.__name__ != "Child":
    print("FAIL: classmethod factory creates wrong class")

print("test_classmethod_staticmethod: PASS")
