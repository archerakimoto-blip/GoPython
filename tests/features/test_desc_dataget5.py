class DataDesc:
    def __get__(self, obj, objtype=None):
        return 25
    
    def __set__(self, obj, value):
        obj._stored = value

class MyObj:
    desc = DataDesc()

m = MyObj()
print(m.desc)
m.desc = 100
print(m._stored)
print(m.desc)
