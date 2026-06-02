class DataDesc:
    def __get__(self, obj, objtype=None):
        return 25
    
    def __set__(self, obj, value):
        print("set called")

class MyObj:
    desc = DataDesc()
    
    def __init__(self):
        self.desc = 100

m = MyObj()
print(m.desc)
