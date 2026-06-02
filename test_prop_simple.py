class Temperature:
    def __init__(self, celsius):
        self._celsius = celsius

    @property
    def celsius(self):
        return self._celsius

    @celsius.setter
    def celsius(self, value):
        self._celsius = value

    @property
    def fahrenheit(self):
        return self._celsius * 9 / 5 + 32

t = Temperature(0)
print(t.celsius)
print(t.fahrenheit)

t.celsius = 100
print(t.celsius)
print(t.fahrenheit)

print("done")
