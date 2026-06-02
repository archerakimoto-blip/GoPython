class DataDesc:
    def __get__(self, obj, objtype=None):
        return 25
    
    def __set__(self, obj, value):
        pass

class MyObj:
    desc = DataDesc()

m = MyObj()
print(m.desc)
m.desc = 100
print(m.desc)
