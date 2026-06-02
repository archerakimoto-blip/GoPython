class Circle:
    def __init__(self, radius):
        self._radius = radius
    
    def get_radius(self):
        return self._radius
    
    def set_radius(self, value):
        self._radius = value
    
    radius = property(get_radius, set_radius)
    
    @classmethod
    def create(cls, radius):
        return cls(radius)
    
    @staticmethod
    def helper(x, y):
        return x + y

c = Circle(5)
print("radius:", c.radius)

c.radius = 10
print("new radius:", c.radius)

c2 = Circle.create(20)
print("created radius:", c2.radius)

print("helper:", Circle.helper(3, 3))

class WithInstance:
    @staticmethod
    def helper(x, y):
        return x * y

wi = WithInstance()
print("instance helper:", wi.helper(4, 2))
