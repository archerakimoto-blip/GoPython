class DataDesc:
    def __get__(self, obj, objtype=None):
        return 25

class MyObj:
    desc = DataDesc()

m = MyObj()
print(m.desc)
