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
if t.celsius != 0:
    print("FAIL: property getter")
if t.fahrenheit != 32:
    print("FAIL: property computed getter")

t.celsius = 100
if t.celsius != 100:
    print("FAIL: property setter")
if t.fahrenheit != 212:
    print("FAIL: property computed after setter")

class ReadOnlyProp:
    @property
    def name(self):
        return "readonly"

ro = ReadOnlyProp()
if ro.name != "readonly":
    print("FAIL: readonly property")
try:
    ro.name = "writable"
    print("FAIL: should not allow setting readonly property")
except:
    pass

class ComputedProp:
    def __init__(self, x):
        self._x = x

    @property
    def doubled(self):
        return self._x * 2

cp = ComputedProp(5)
if cp.doubled != 10:
    print("FAIL: computed property")

cp._x = 7
if cp.doubled != 14:
    print("FAIL: computed property after internal change")

print("test_property: PASS")
