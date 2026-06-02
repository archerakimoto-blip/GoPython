class Point:
    __slots__ = ["x", "y"]
    
    def __init__(self, x, y):
        self.x = x
        self.y = y

p = Point(3, 4)
print(p.x)
print(p.y)

p.x = 10
print(p.x)

try:
    p.z = 5
except:
    print("cannot set z")

class ColorPoint(Point):
    __slots__ = ["color"]
    
    def __init__(self, x, y, color):
        self.x = x
        self.y = y
        self.color = color

cp = ColorPoint(1, 2, "red")
print(cp.x)
print(cp.color)

try:
    cp.extra = 99
except:
    print("cannot set extra on ColorPoint")

class FreeChild(Point):
    def __init__(self, x, y):
        self.x = x
        self.y = y

fc = FreeChild(7, 8)
fc.z = 100
print(fc.z)

class Empty:
    __slots__ = []
    
    def __init__(self):
        pass

e = Empty()
try:
    e.foo = 1
except:
    print("cannot set foo on Empty")

class SlotWithMethod:
    __slots__ = ["name"]
    
    def __init__(self, name):
        self.name = name
    
    def greet(self):
        return "hello " + self.name

sm = SlotWithMethod("world")
print(sm.greet())
print(sm.name)
