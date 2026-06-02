class Circle:
    def __init__(self, radius):
        self._radius = radius
    
    def get_radius(self):
        return self._radius
    
    def set_radius(self, value):
        self._radius = value
    
    radius = property(get_radius, set_radius)

c = Circle(5)
print(c.radius)
c.radius = 10
print(c.radius)
