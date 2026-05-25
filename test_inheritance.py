# Test class inheritance and polymorphism

class Animal {
    def speak(self):
        return "Animal speaks"
    
    def move(self):
        return "Animal moves"
}

class Dog(Animal) {
    def speak(self):
        return "Woof!"
}

class Cat(Animal) {
    def speak(self):
        return "Meow!"
}

# Create instances
animal = Animal()
dog = Dog()
cat = Cat()

# Test basic inheritance
print("Animal speak:", animal.speak())
print("Dog speak:", dog.speak())
print("Cat speak:", cat.speak())

# Test inherited methods
print("Dog move:", dog.move())
print("Cat move:", cat.move())
