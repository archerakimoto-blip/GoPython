
class Temperature:
    def __init__(self, celsius):
        self._celsius = celsius
        self._deleted = false

    @property
    def celsius(self):
        return self._celsius

    @celsius.setter
    def celsius(self, value):
        self._celsius = value

    @celsius.deleter
    def celsius(self):
        self._deleted = true

t = Temperature(100)
print(t.celsius)
t.celsius = 200
print(t.celsius)
del t.celsius
print(t._deleted)

# Test del on simple instance attribute
class Simple:
    def __init__(self):
        self.x = 10
        self.y = 20

s = Simple()
print(s.x)
del s.x
print(s.y)
