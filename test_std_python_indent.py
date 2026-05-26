class Animal:
    def speak(self):
        return "Animal sound"

class Dog(Animal):
    def speak(self):
        return "Woof!"

a = Animal()
print(a.speak())

d = Dog()
print(d.speak())

def add(x, y):
    return x + y

result = add(3, 5)
print(result)
