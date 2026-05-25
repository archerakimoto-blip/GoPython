class Animal: {
    def speak(self): {
        return "Animal sound"
    }
}

class Dog(Animal): {
    def speak(self): {
        return "Woof!"
    }
}

a = Animal()
print(a.speak())

d = Dog()
print(d.speak())
