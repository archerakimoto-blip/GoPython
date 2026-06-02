class Desc:
    def __get__(self, obj, objtype=None):
        return "desc_value"
    
    def __set__(self, obj, value):
        print("set called")

class MyClass:
    x = Desc()

print(MyClass.x)
m = MyClass()
print(m.x)
m.x = 999
print(m.x)
