class NonDataDesc:
    def __get__(self, obj, objtype=None):
        return "descriptor_value_for_prop"

class Prop:
    desc = NonDataDesc()
    
    def __init__(self):
        self.desc = "instance_value"

p = Prop()
print(p.desc)
