class NonDataDesc:
    def __get__(self, obj, objtype=None):
        return "descriptor_value_for_prop"

class Prop:
    desc = NonDataDesc()
    
    def __init__(self):
        self.desc = "instance_value"

p = Prop()
print(p.desc)

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
print(m.desc)

class ReadOnlyDesc:
    def __get__(self, obj, objtype=None):
        return "readonly_value"
    
    def __set__(self, obj, value):
        raise Exception("cannot set readonly attribute")

class ReadOnly:
    x = ReadOnlyDesc()

r = ReadOnly()
print(r.x)
try:
    r.x = 999
except:
    print("cannot set readonly attribute")
print(r.x)
