class DataDesc:
    def __get__(self, obj, objtype=None):
        return 25
    
    def __set__(self, obj, value):
        pass

class MyObj:
    desc = DataDesc()
    
    def __init__(self):
        self.desc = 100

m = MyObj()
print("m type:", m)
print("m.desc:")
print(m.desc)
