class DataDesc:
    def __get__(self, obj, objtype=None):
        return 25
    
    def __set__(self, obj, value):
        pass

class MyObj:
    desc = DataDesc()
    
    def __init__(self):
        self.x = 1
        self.desc = 100
        self.y = 2

m = MyObj()
print(m.x)
print(m.y)
print(m.desc)
