class DataDesc:
    def __set__(self, obj, value):
        pass

class MyObj:
    desc = DataDesc()
    
    def __init__(self):
        self.desc = 100

m = MyObj()
print(m.desc)
