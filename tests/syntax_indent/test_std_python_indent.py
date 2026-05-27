class Animal:
    def speak(self):
        return "Animal sound"

class Dog(Animal):
    def speak(self):
        return "Woof!"

a = Animal()
d = Dog()
print(a.speak())
print(d.speak())