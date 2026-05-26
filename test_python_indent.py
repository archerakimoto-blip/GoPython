class Animal:
    def speak(self):
        return "Animal sound"

class Dog(Animal):
    def speak(self):
        return "Woof!"

class Cat(Animal):
    def speak(self):
        return "Meow!"

a = Animal()
d = Dog()
c = Cat()
print(a.speak())
print(d.speak())
print(c.speak())