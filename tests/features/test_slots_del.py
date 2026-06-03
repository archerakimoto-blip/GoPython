class Point:
    __slots__ = ["x", "y"]
    
    def __init__(self, x, y):
        self.x = x
        self.y = y

p = Point(3, 4)
print(p.x)

del p.x
try:
    v = p.x
    print("FAIL: should not be able to access deleted slot, got: " + str(v))
except:
    print("deleted slot inaccessible")

p.x = 99
print(p.x)

try:
    del p.z
    print("FAIL: should not be able to delete non-slot")
except:
    print("cannot delete non-slot attribute")

print("test_slots_del: PASS")
