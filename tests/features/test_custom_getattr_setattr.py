class WithProperty:
    def __init__(self):
        self._val = 0

    @property
    def val(self):
        return self._val

    @val.setter
    def val(self, v):
        self._val = v

wp = WithProperty()
if wp.val != 0:
    print("FAIL: property initial")
wp.val = 5
if wp.val != 5:
    print("FAIL: property set")

class MixedAccess(WithProperty):
    def __getattr__(self, name):
        if name == "extra":
            return 99
        return "mixed_" + name

ma = MixedAccess()
if ma.val != 0:
    print("FAIL: mixed property")
ma.val = 7
if ma.val != 7:
    print("FAIL: mixed property set")
if ma.extra != 99:
    print("FAIL: mixed __getattr__")
if ma.unknown != "mixed_unknown":
    print("FAIL: mixed __getattr__ default")

print("test_custom_getattr_setattr: PASS")
