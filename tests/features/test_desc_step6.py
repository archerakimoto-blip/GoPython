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
print(c.radius)
c.radius = 10
print(c.radius)

c2 = Circle.create(20)
print(c2.radius)

print(Circle.helper(3, 3))
