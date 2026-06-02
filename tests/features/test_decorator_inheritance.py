class Animal:
    def __init__(self, name):
        self.name = name

    @property
    def sound(self):
        return self._sound

    @sound.setter
    def sound(self, value):
        self._sound = value

    def speak(self):
        return self.name + " says " + self.sound

class Dog(Animal):
    def __init__(self, name):
        super().__init__(name)
        self.sound = "woof"

class Cat(Animal):
    def __init__(self, name):
        super().__init__(name)
        self.sound = "meow"

    @classmethod
    def create(cls, name):
        return cls(name)

    @staticmethod
    def is_feline():
        return "yes"

d = Dog("Rex")
if d.speak() != "Rex says woof":
    print("FAIL: property inheritance")

c = Cat("Whiskers")
if c.speak() != "Whiskers says meow":
    print("FAIL: property inheritance with classmethod")
if Cat.is_feline() != "yes":
    print("FAIL: staticmethod inheritance")

c2 = Cat.create("Tom")
if c2.name != "Tom":
    print("FAIL: classmethod factory")
if c2.speak() != "Tom says meow":
    print("FAIL: classmethod created instance")

class Base:
    @property
    def value(self):
        return 10

class Middle(Base):
    pass

class Child(Middle):
    @property
    def value(self):
        return 20

b = Base()
m = Middle()
ch = Child()
if b.value != 10:
    print("FAIL: base property")
if m.value != 10:
    print("FAIL: inherited property")
if ch.value != 20:
    print("FAIL: overridden property")

print("test_decorator_inheritance: PASS")
