# GoPy Runtime Test Suite
# Comprehensive test for VM behavior - compatible with GoPy parser limitations
# No multi-arg print, no tuple literals, no dict literals, no f-strings, no walrus, no decorators

_passed = 0
_failed = 0
_total = 0

def check(name, actual, expected):
    global _passed, _failed, _total
    _total = _total + 1
    if actual == expected:
        _passed = _passed + 1
    else:
        _failed = _failed + 1
        print("FAIL: " + name + " => got " + str(actual) + " want " + str(expected))

def check_true(name, cond):
    global _passed, _failed, _total
    _total = _total + 1
    if cond:
        _passed = _passed + 1
    else:
        _failed = _failed + 1
        print("FAIL: " + name)

# ============================================================
# Section 1: Basic Types and Arithmetic
# ============================================================

check("int add", 1 + 2, 3)
check("int sub", 10 - 3, 7)
check("int mul", 4 * 5, 20)
check("int div", 10 // 3, 3)
check("int mod", 10 % 3, 1)
check("int pow", 2 ** 10, 1024)
check("int neg", -5, -5)
check("float add", 1.5 + 2.5, 4.0)
check("float sub", 5.5 - 1.5, 4.0)
check("float mul", 2.5 * 4.0, 10.0)
check("float div", 7.0 / 2.0, 3.5)
check("int float mix", 1 + 2.0, 3.0)

# ============================================================
# Section 2: Comparison and Boolean
# ============================================================

check_true("eq", 1 == 1)
check_true("ne", 1 != 2)
check_true("lt", 1 < 2)
check_true("gt", 2 > 1)
check_true("le", 1 <= 1)
check_true("ge", 2 >= 1)
check_true("and true", True and True)
check_true("or true", True or False)
check_true("not false", not False)
check_true("and false", not (True and False))
check_true("or false", not (False or False))
check_true("chained cmp", 1 < 2 < 3)
check_true("chained cmp2", not (1 < 2 < 1))

# ============================================================
# Section 3: String Operations
# ============================================================

check("str concat", "hello" + " " + "world", "hello world")
check("str repeat", "ab" * 3, "ababab")
check("str len", len("hello"), 5)
check("str index", "hello"[1], "e")
check("str slice", "hello"[1:4], "ell")
check("str neg index", "hello"[-1], "o")
check_true("str in", "ell" in "hello")
check_true("str not in", "xyz" not in "hello")
check("str upper", "Hello".upper(), "HELLO")
check("str lower", "Hello".lower(), "hello")
check("str strip", "  hi  ".strip(), "hi")
check("str lstrip", "  hi  ".lstrip(), "hi  ")
check("str rstrip", "  hi  ".rstrip(), "  hi")
check("str replace", "hello world".replace("world", "python"), "hello python")
check("str split len", len("a,b,c".split(",")), 3)
check("str join", "-".join(["a", "b", "c"]), "a-b-c")
check("str startswith", "hello".startswith("he"), True)
check("str endswith", "hello".endswith("lo"), True)
check("str find", "hello".find("ll"), 2)
check("str find miss", "hello".find("xx"), -1)
check("str count", "hello".count("l"), 2)
check("str format", "Hello, {}!".format("World"), "Hello, World!")
check("str isdigit", "123".isdigit(), True)
check("str isalpha", "abc".isalpha(), True)

# ============================================================
# Section 4: List Operations
# ============================================================

lst = [1, 2, 3]
check("list index", lst[0], 1)
check("list neg index", lst[-1], 3)
check("list len", len(lst), 3)
lst.append(4)
check("list append", len(lst), 4)
lst.extend([5, 6])
check("list extend", len(lst), 6)
lst.insert(0, 0)
check("list insert", lst[0], 0)
v = lst.pop()
check("list pop val", v, 6)
check("list pop len", len(lst), 6)
lst.remove(0)
check("list remove", lst[0], 1)
check("list index of", lst.index(3), 2)
check("list count", [1, 2, 2, 3].count(2), 2)
lst2 = [3, 1, 2]
lst2.sort()
check("list sort", lst2[0], 1)
check("list sort2", lst2[1], 2)
lst3 = [1, 2, 3]
lst3.reverse()
check("list reverse", lst3[0], 3)
check("list concat", [1, 2] + [3, 4], [1, 2, 3, 4])
check("list repeat", [0] * 3, [0, 0, 0])
check("list slice", [0, 1, 2, 3, 4][1:3], [1, 2])
check_true("list in", 2 in [1, 2, 3])
check_true("list not in", 4 not in [1, 2, 3])

# ============================================================
# Section 5: Dict Operations
# ============================================================

d = {}
d["a"] = 1
d["b"] = 2
check("dict setget", d["a"], 1)
check("dict len", len(d), 2)
check("dict get", d.get("a"), 1)
check("dict get default", d.get("x", 0), 0)
check_true("dict in", "a" in d)
check_true("dict not in", "z" not in d)
d["c"] = 3
check("dict assign", len(d), 3)
v = d.pop("a")
check("dict pop", v, 1)
check("dict pop len", len(d), 2)
d.update({"x": 10})
check("dict update", d["x"], 10)
d.setdefault("y", 20)
check("dict setdefault", d["y"], 20)
check("dict keys len", len(d.keys()), 4)
check("dict values len", len(d.values()), 4)
check("dict items len", len(d.items()), 4)

# Dict merge
d1 = {}
d1["a"] = 1
d2 = {}
d2["b"] = 2
d3 = d1 | d2
check("dict merge len", len(d3), 2)
check("dict merge a", d3["a"], 1)
check("dict merge b", d3["b"], 2)

# ============================================================
# Section 6: Set Operations
# ============================================================

s = set()
s.add(1)
s.add(2)
s.add(3)
check("set add len", len(s), 3)
check_true("set in", 1 in s)
s.add(2)
check("set add dup", len(s), 3)
s.remove(1)
check("set remove", len(s), 2)
check_true("set not in", 1 not in s)

# ============================================================
# Section 7: Variables and Scope
# ============================================================

x = 10
check("global var", x, 10)
x = x + 5
check("global reassign", x, 15)
x += 3
check("augmented +=", x, 18)
x -= 2
check("augmented -=", x, 16)
x *= 2
check("augmented *=", x, 32)
x //= 3
check("augmented //=", x, 10)
x %= 3
check("augmented %=", x, 1)

# ============================================================
# Section 8: Control Flow
# ============================================================

# if/elif/else
def classify(n):
    if n < 0:
        return "neg"
    elif n == 0:
        return "zero"
    else:
        return "pos"
check("if elif", classify(-1), "neg")
check("if else", classify(0), "zero")
check("if elif2", classify(1), "pos")

# for loop
total = 0
for i in range(5):
    total = total + i
check("for range", total, 10)

# while loop
n = 0
while n < 5:
    n = n + 1
check("while", n, 5)

# break
result = []
for i in range(10):
    if i == 5:
        break
    result.append(i)
check("break len", len(result), 5)
check("break last", result[-1], 4)

# continue
result2 = []
for i in range(10):
    if i % 2 == 0:
        continue
    result2.append(i)
check("continue len", len(result2), 5)
check("continue first", result2[0], 1)
check("continue last", result2[-1], 9)

# nested loop
count = 0
for i in range(3):
    for j in range(3):
        count = count + 1
check("nested loop", count, 9)

# for else
def has_item(lst, target):
    for x in lst:
        if x == target:
            return True
    else:
        return False
check("for else found", has_item([1, 2, 3], 2), True)
check("for else not found", has_item([1, 2, 3], 5), False)

# ============================================================
# Section 9: Functions
# ============================================================

def add(a, b):
    return a + b
check("func basic", add(3, 4), 7)

def greet(name, greeting="hello"):
    return greeting + " " + name
check("func default", greet("world"), "hello world")
check("func default override", greet("world", "hi"), "hi world")

def multi_default(a, b=2, c=3):
    return a + b + c
check("func multi default", multi_default(1), 6)
check("func multi default2", multi_default(1, 5), 9)
check("func multi default3", multi_default(1, 5, 7), 13)

def sum_all(*args):
    total = 0
    for x in args:
        total = total + x
    return total
check("func *args", sum_all(1, 2, 3), 6)
check("func *args2", sum_all(1, 2, 3, 4, 5), 15)

def get_kw(**kwargs):
    return kwargs
r = get_kw(a=1, b=2)
check("func **kwargs a", r["a"], 1)
check("func **kwargs b", r["b"], 2)

def mixed(a, b, *args, **kwargs):
    return a + b + sum_all(*args)
check("func mixed", mixed(1, 2, 3, 4), 10)

# Recursive
def factorial(n):
    if n <= 1:
        return 1
    return n * factorial(n - 1)
check("factorial 5", factorial(5), 120)
check("factorial 10", factorial(10), 3628800)

def fib(n):
    if n <= 1:
        return n
    return fib(n - 1) + fib(n - 2)
check("fib 10", fib(10), 55)

# ============================================================
# Section 10: Closures
# ============================================================

def make_adder(n):
    def adder(x):
        return x + n
    return adder
add5 = make_adder(5)
add10 = make_adder(10)
check("closure add5", add5(3), 8)
check("closure add10", add10(3), 13)

def make_counter():
    count = 0
    def counter():
        nonlocal count
        count = count + 1
        return count
    return counter
c = make_counter()
check("counter 1", c(), 1)
check("counter 2", c(), 2)
check("counter 3", c(), 3)

def outer(x):
    def middle(y):
        def inner(z):
            return x + y + z
        return inner
    return middle
check("nested closure", outer(1)(2)(3), 6)

# Closure capturing loop variable
def make_funcs():
    funcs = []
    for i in range(3):
        def make_func(val):
            def func():
                return val
            return func
        funcs.append(make_func(i))
    check("closure loop 0", funcs[0](), 0)
    check("closure loop 1", funcs[1](), 1)
    check("closure loop 2", funcs[2](), 2)
make_funcs()

# ============================================================
# Section 11: Classes
# ============================================================

class Dog:
    def __init__(self, name):
        self.name = name
    def speak(self):
        return self.name + " says woof"

d = Dog("Rex")
check("class attr", d.name, "Rex")
check("class method", d.speak(), "Rex says woof")

class Counter:
    def __init__(self):
        self.value = 0
    def increment(self):
        self.value = self.value + 1
        return self.value
    def get(self):
        return self.value

ctr = Counter()
check("class incr 1", ctr.increment(), 1)
check("class incr 2", ctr.increment(), 2)
check("class get", ctr.get(), 2)

# Inheritance
class Animal:
    def __init__(self, name):
        self.name = name
    def speak(self):
        return "..."

class Cat(Animal):
    def speak(self):
        return self.name + " says meow"

cat = Cat("Whiskers")
check("inherit attr", cat.name, "Whiskers")
check("inherit method", cat.speak(), "Whiskers says meow")

# Class attribute
class Tracker:
    count = 0
    def __init__(self):
        Tracker.count = Tracker.count + 1

t1 = Tracker()
t2 = Tracker()
check("class attr", Tracker.count, 2)

# __str__ method
class Wrapper:
    def __init__(self, val):
        self.val = val
    def __str__(self):
        return "W:" + str(self.val)

w = Wrapper(42)
check("class str", str(w), "W:42")

# ============================================================
# Section 12: Exception Handling
# ============================================================

def safe_div(a, b):
    try:
        return a / b
    except:
        return -1
check("try except div", safe_div(10, 2), 5.0)
check("try except zero", safe_div(10, 0), -1)

def check_value(x):
    if x < 0:
        raise ValueError("negative")
    return True

check("no raise", check_value(5), True)
caught = False
try:
    check_value(-1)
except ValueError:
    caught = True
check("raise catch", caught, True)

# try/finally
finally_ran = False
try:
    x = 1 + 1
finally:
    finally_ran = True
check("finally runs", finally_ran, True)

# nested try
def nested_try():
    try:
        try:
            raise ValueError("inner")
        except ValueError:
            return "caught inner"
    except:
        return "caught outer"
check("nested try", nested_try(), "caught inner")

# ============================================================
# Section 13: Built-in Functions
# ============================================================

check("len list", len([1, 2, 3]), 3)
check("len str", len("hello"), 5)
check("len dict", len(d), 4)
check("abs", abs(-5), 5)
check("abs pos", abs(3), 3)
check("min list", min([3, 1, 2]), 1)
check("max list", max([3, 1, 2]), 3)
check("sum list", sum([1, 2, 3, 4, 5]), 15)
check("sorted", sorted([3, 1, 2]), [1, 2, 3])
check("reversed", list(reversed([1, 2, 3])), [3, 2, 1])
check("range list", list(range(5)), [0, 1, 2, 3, 4])
check("range step", list(range(0, 10, 2)), [0, 2, 4, 6, 8])
check("chr", chr(65), "A")
check("ord", ord("A"), 65)
check("hex", hex(255), "0xff")
check("oct", oct(8), "0o10")
check("bin", bin(10), "0b1010")
check("str int", str(123), "123")
check("int str", int("42"), 42)
check("float str", float("3.14"), 3.14)
check("bool true", bool(1), True)
check("bool false", bool(0), False)
check("type int", type(1).__name__, "int")
check("type str", type("a").__name__, "str")
check("isinstance int", isinstance(1, int), True)
check("isinstance str", isinstance("a", str), True)

# map/filter
doubled = list(map(lambda x: x * 2, [1, 2, 3]))
check("map", doubled, [2, 4, 6])
evens = list(filter(lambda x: x % 2 == 0, [1, 2, 3, 4, 5]))
check("filter", evens, [2, 4])

# zip
zipped = list(zip([1, 2, 3], ["a", "b", "c"]))
check("zip len", len(zipped), 3)
check("zip first0", zipped[0][0], 1)
check("zip first1", zipped[0][1], "a")

# enumerate
enumed = list(enumerate(["a", "b", "c"]))
check("enum len", len(enumed), 3)
check("enum idx", enumed[0][0], 0)
check("enum val", enumed[0][1], "a")

# any/all
check("any true", any([False, True, False]), True)
check("any false", any([False, False, False]), False)
check("all true", all([True, True, True]), True)
check("all false", all([True, False, True]), False)

# ============================================================
# Section 14: Comprehensions
# ============================================================

squares = [x * x for x in range(5)]
check("list comp", squares, [0, 1, 4, 9, 16])

evens_comp = [x for x in range(10) if x % 2 == 0]
check("list comp filter", evens_comp, [0, 2, 4, 6, 8])

# ============================================================
# Section 15: Lambda
# ============================================================

f = lambda x: x * 2
check("lambda basic", f(5), 10)
g = lambda x, y: x + y
check("lambda multi", g(2, 3), 5)

# ============================================================
# Section 16: Generator
# ============================================================

def gen(n):
    for i in range(n):
        yield i
check("generator", list(gen(5)), [0, 1, 2, 3, 4])

def gen_even(n):
    for i in range(n):
        if i % 2 == 0:
            yield i
check("generator filter", list(gen_even(6)), [0, 2, 4])

# ============================================================
# Section 17: Ternary Expression
# ============================================================

x = 5
check("ternary true", "big" if x > 3 else "small", "big")
check("ternary false", "big" if x < 3 else "small", "small")

# ============================================================
# Section 18: Multiple Assignment
# ============================================================

a, b = 10, 20
check("multi assign a", a, 10)
check("multi assign b", b, 20)

# ============================================================
# Section 19: Slicing
# ============================================================

lst = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
check("slice basic", lst[2:5], [2, 3, 4])
check("slice step", lst[::2], [0, 2, 4, 6, 8])
check("slice neg step", lst[::-1], [9, 8, 7, 6, 5, 4, 3, 2, 1, 0])
check("slice neg idx", lst[-3:], [7, 8, 9])

# ============================================================
# Section 20: Nested Data Structures
# ============================================================

matrix = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
check("nested access", matrix[1][2], 6)
check("nested row len", len(matrix[0]), 3)

# ============================================================
# Section 21: String Formatting
# ============================================================

check("format basic", "Hello, {}!".format("World"), "Hello, World!")
check("format multi", "{} + {} = {}".format(1, 2, 3), "1 + 2 = 3")

# ============================================================
# Section 22: Set Operations
# ============================================================

s1 = set()
s1.add(1)
s1.add(2)
s1.add(3)
s2 = set()
s2.add(2)
s2.add(3)
s2.add(4)
su = s1 | s2
check("set union len", len(su), 4)
si = s1 & s2
check("set intersect len", len(si), 2)
sd = s1 - s2
check("set diff len", len(sd), 1)

# ============================================================
# Section 23: None handling
# ============================================================

x = None
check_true("is none", x is None)
check_true("is not none", 1 is not None)

# ============================================================
# Section 24: Bitwise Operations
# ============================================================

check("bit and", 5 & 3, 1)
check("bit or", 5 | 3, 7)
check("bit xor", 5 ^ 3, 6)
check("bit not", ~0, -1)
check("bit lshift", 1 << 4, 16)
check("bit rshift", 16 >> 2, 4)

# ============================================================
# Section 25: Complex Expressions
# ============================================================

check("complex expr", (1 + 2) * (3 + 4), 21)
check("nested call", len(str(12345)), 5)
check("chained method", "  hello  ".strip().upper(), "HELLO")

# ============================================================
# Section 26: Higher-order Functions
# ============================================================

def apply(f, x):
    return f(x)
check("hof", apply(lambda x: x * 2, 5), 10)

def compose(f, g):
    def composed(x):
        return f(g(x))
    return composed
double = lambda x: x * 2
inc = lambda x: x + 1
check("compose", compose(double, inc)(5), 12)

# ============================================================
# Section 27: Decorator Pattern (manual)
# ============================================================

def double_result(func):
    def wrapper(x):
        return func(x) * 2
    return wrapper

def square(x):
    return x * x
square_d = double_result(square)
check("decorator manual", square_d(3), 18)

# ============================================================
# Section 28: Context Manager (with)
# ============================================================

class MyCtx:
    def __init__(self):
        self.entered = False
        self.exited = False
    def __enter__(self):
        self.entered = True
        return self
    def __exit__(self, *args):
        self.exited = True
        return False

ctx = MyCtx()
with ctx as c:
    check("ctx entered", c.entered, True)
    check("ctx not exited", c.exited, False)
check("ctx exited", ctx.exited, True)

# ============================================================
# Section 29: Property-like access
# ============================================================

class Rect:
    def __init__(self, w, h):
        self.w = w
        self.h = h
    def area(self):
        return self.w * self.h
    def perimeter(self):
        return 2 * (self.w + self.h)

r = Rect(3, 4)
check("rect area", r.area(), 12)
check("rect perimeter", r.perimeter(), 14)

# ============================================================
# Section 30: String in/out
# ============================================================

check("str bool", str(True), "True")
check("str none", str(None), "None")
check("str list", str([1, 2, 3]), "[1, 2, 3]")

# ============================================================
# Section 31: Multiple return values via list
# ============================================================

def divmod_func(a, b):
    return [a // b, a % b]
r = divmod_func(17, 5)
check("divmod q", r[0], 3)
check("divmod r", r[1], 2)

# ============================================================
# Section 32: Dynamic attribute setting
# ============================================================

class Empty:
    pass
e = Empty()
e.x = 10
e.y = 20
check("dynamic attr x", e.x, 10)
check("dynamic attr y", e.y, 20)

# ============================================================
# Section 33: Method chaining
# ============================================================

class Builder:
    def __init__(self):
        self.result = ""
    def add(self, s):
        self.result = self.result + s
        return self
    def build(self):
        return self.result

check("chain", Builder().add("a").add("b").add("c").build(), "abc")

# ============================================================
# Section 34: Nested function scope
# ============================================================

def outer_fn():
    x = 10
    def inner_fn():
        return x
    return inner_fn()
check("nested scope", outer_fn(), 10)

# ============================================================
# Section 35: Pass statement
# ============================================================

class Abstract:
    pass
a = Abstract()
check("pass class", type(a).__name__, "Abstract")

# ============================================================
# Section 36: Del statement
# ============================================================

deld = {}
deld["x"] = 1
deld["y"] = 2
del deld["x"]
check("del key", len(deld), 1)
check_true("del key not in", "x" not in deld)

# ============================================================
# Section 37: Assert statement
# ============================================================

try:
    assert True
    assert_result = "pass"
except:
    assert_result = "fail"
check("assert true", assert_result, "pass")

# ============================================================
# Section 38: Global statement
# ============================================================

gvar = 0
def modify_global():
    global gvar
    gvar = 42
modify_global()
check("global", gvar, 42)

# ============================================================
# Section 39: Complex class interactions
# ============================================================

class Vec:
    def __init__(self, x, y):
        self.x = x
        self.y = y
    def add(self, other):
        return Vec(self.x + other.x, self.y + other.y)
    def scale(self, s):
        return Vec(self.x * s, self.y * s)
    def mag_sq(self):
        return self.x * self.x + self.y * self.y

v1 = Vec(3, 4)
v2 = Vec(1, 2)
v3 = v1.add(v2)
check("vec add x", v3.x, 4)
check("vec add y", v3.y, 6)
v4 = v1.scale(2)
check("vec scale x", v4.x, 6)
check("vec scale y", v4.y, 8)
check("vec mag_sq", v1.mag_sq(), 25)

# ============================================================
# Section 40: List comprehension with method calls
# ============================================================

words = ["hello", "world", "python"]
upper = [w.upper() for w in words]
check("comp method", upper, ["HELLO", "WORLD", "PYTHON"])

# ============================================================
# Section 41: Nested comprehension
# ============================================================

flat = [x for row in [[1, 2], [3, 4], [5, 6]] for x in row]
check("nested comp", flat, [1, 2, 3, 4, 5, 6])

# ============================================================
# Section 42: Generator expression
# ============================================================

total = sum([x * x for x in range(5)])
check("gen expr sum", total, 30)

# ============================================================
# Section 43: String methods advanced
# ============================================================

check("str capitalize", "hello world".capitalize(), "Hello world")
check("str title", "hello world".title(), "Hello World")
check("str swapcase", "Hello".swapcase(), "hELLO")
check("str center", "hi".center(6), "  hi  ")
check("str ljust", "hi".ljust(5), "hi   ")
check("str rjust", "hi".rjust(5), "   hi")
check("str zfill", "42".zfill(5), "00042")
check("str partition", "hello world".partition(" "), ("hello", " ", "world"))

# ============================================================
# Section 44: List copy
# ============================================================

orig = [1, 2, 3]
copy = orig.copy()
copy.append(4)
check("list copy orig", len(orig), 3)
check("list copy copy", len(copy), 4)

# ============================================================
# Section 45: Dict fromkeys
# ============================================================

d4 = dict.fromkeys(["a", "b", "c"], 0)
check("fromkeys len", len(d4), 3)
check("fromkeys val", d4["a"], 0)

# ============================================================
# Section 46: isinstance with custom class
# ============================================================

class Base:
    pass
class Derived(Base):
    pass
b = Base()
d_obj = Derived()
check("isinstance base", isinstance(b, Base), True)
check("isinstance derived", isinstance(d_obj, Base), True)
check("isinstance not", isinstance(b, Derived), False)

# ============================================================
# Section 47: Exception hierarchy
# ============================================================

try:
    raise TypeError("type err")
except TypeError:
    caught_type = True
except:
    caught_type = False
check("catch TypeError", caught_type, True)

try:
    raise ValueError("val err")
except ValueError:
    caught_val = True
except:
    caught_val = False
check("catch ValueError", caught_val, True)

# ============================================================
# Section 48: Multiple except handlers
# ============================================================

def multi_except(err_type):
    try:
        if err_type == "value":
            raise ValueError("v")
        elif err_type == "type":
            raise TypeError("t")
        else:
            raise KeyError("k")
    except ValueError:
        return "ValueError"
    except TypeError:
        return "TypeError"
    except:
        return "other"
check("multi except value", multi_except("value"), "ValueError")
check("multi except type", multi_except("type"), "TypeError")
check("multi except other", multi_except("key"), "other")

# ============================================================
# Section 49: Chained method calls on builtins
# ============================================================

check("chained str", "  HELLO  ".strip().lower(), "hello")
check("chained list", sorted([3,1,2]).__len__(), 3)

# ============================================================
# Section 50: Complex real-world patterns
# ============================================================

# Stack implementation
class Stack:
    def __init__(self):
        self.items = []
    def push(self, item):
        self.items.append(item)
    def pop(self):
        return self.items.pop()
    def is_empty(self):
        return len(self.items) == 0
    def size(self):
        return len(self.items)

s = Stack()
s.push(1)
s.push(2)
s.push(3)
check("stack pop", s.pop(), 3)
check("stack size", s.size(), 2)
check("stack not empty", s.is_empty(), False)

# Fibonacci with memoization
def make_fib():
    cache = {}
    def fib(n):
        if n in cache:
            return cache[n]
        if n <= 1:
            result = n
        else:
            result = fib(n - 1) + fib(n - 2)
        cache[n] = result
        return result
    return fib
fast_fib = make_fib()
check("memo fib 20", fast_fib(20), 6765)

# ============================================================
# Results
# ============================================================

print("=" * 60)
print("Results: " + str(_passed) + " passed, " + str(_failed) + " failed, " + str(_total) + " total")
if _failed == 0:
    print("ALL TESTS PASSED")
else:
    print("SOME TESTS FAILED")
