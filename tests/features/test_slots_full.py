class Point:
    __slots__ = ["x", "y"]

    def __init__(self, x, y):
        self.x = x
        self.y = y

    def sum(self):
        return self.x + self.y

p = Point(3, 4)
if p.x != 3:
    print("FAIL: slots basic get x")
if p.y != 4:
    print("FAIL: slots basic get y")
if p.sum() != 7:
    print("FAIL: slots method call")

p.x = 10
if p.x != 10:
    print("FAIL: slots set")

try:
    p.z = 5
    print("FAIL: should not allow setting non-slot attribute")
except:
    pass

class ColorPoint(Point):
    __slots__ = ["color"]

    def __init__(self, x, y, color):
        self.x = x
        self.y = y
        self.color = color

cp = ColorPoint(1, 2, "red")
if cp.x != 1:
    print("FAIL: slots inheritance get x")
if cp.y != 2:
    print("FAIL: slots inheritance get y")

try:
    cp.extra = 99
    print("FAIL: should not allow setting non-slot on child")
except:
    pass

class FreeChild(Point):
    def __init__(self, x, y):
        self.x = x
        self.y = y

fc = FreeChild(7, 8)
fc.z = 100
if fc.z != 100:
    print("FAIL: free child should allow new attributes")

print("test_slots: PASS")
