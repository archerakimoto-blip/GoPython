#!/usr/bin/env python3
"""GoPy Runtime Test Suite - Comprehensive test for VM behavior comparison"""

import sys

# Test counter
_passed = 0
_failed = 0
_tests = []

def test(name):
    """Decorator to register a test"""
    def decorator(func):
        _tests.append((name, func))
        return func
    return decorator

def run_tests():
    """Run all registered tests"""
    global _passed, _failed
    print(f"Running {_len(_tests)} tests...")
    print("=" * 60)
    for name, func in _tests:
        try:
            func()
            print(f"[PASS] {name}")
            _passed += 1
        except Exception as e:
            print(f"[FAIL] {name}: {e}")
            _failed += 1
    print("=" * 60)
    print(f"Results: {_passed} passed, {_failed} failed")
    return _failed == 0

# Helper
def _len(x):
    return x.__len__() if hasattr(x, '__len__') else 0

# ========== Basic Types ==========

@test("int arithmetic")
def test_int_arithmetic():
    a = 10
    b = 3
    assert a + b == 13
    assert a - b == 7
    assert a * b == 30
    assert a // b == 3
    assert a % b == 1
    assert a ** b == 1000

@test("float arithmetic")
def test_float_arithmetic():
    a = 10.5
    b = 2.5
    assert a + b == 13.0
    assert a - b == 8.0
    assert a * b == 26.25

@test("bool operations")
def test_bool_operations():
    assert True and False == False
    assert True or False == True
    assert not True == False
    assert True and True == True

@test("string operations")
def test_string_operations():
    s = "hello"
    assert s + " world" == "hello world"
    assert s * 2 == "hellohello"
    assert len(s) == 5
    assert s[0] == "h"
    assert s[1:4] == "ell"
    assert "ell" in s

@test("list operations")
def test_list_operations():
    lst = [1, 2, 3]
    lst.append(4)
    assert len(lst) == 4
    assert lst[0] == 1
    assert lst[-1] == 4
    lst2 = lst + [5, 6]
    assert len(lst2) == 6

@test("dict operations")
def test_dict_operations():
    d = {"a": 1, "b": 2}
    assert d["a"] == 1
    d["c"] = 3
    assert len(d) == 3
    assert "a" in d
    assert d.get("x", 0) == 0

@test("set operations")
def test_set_operations():
    s = {1, 2, 3}
    s.add(4)
    assert len(s) == 4
    assert 1 in s
    s2 = s | {5, 6}
    assert len(s2) == 6

# ========== Functions ==========

@test("function definition and call")
def test_function_basic():
    def add(a, b):
        return a + b
    assert add(2, 3) == 5
    assert add(10, 20) == 30

@test("default arguments")
def test_default_args():
    def greet(name, greeting="hello"):
        return greeting + " " + name
    assert greet("world") == "hello world"
    assert greet("world", "hi") == "hi world"

@test("multiple default arguments")
def test_multiple_default_args():
    def func(a, b=2, c=3):
        return a + b + c
    assert func(1) == 6
    assert func(1, 5) == 9
    assert func(1, 5, 7) == 13

@test("*args")
def test_varargs():
    def sum_all(*args):
        total = 0
        for x in args:
            total = total + x
        return total
    assert sum_all(1, 2, 3) == 6
    assert sum_all(1, 2, 3, 4, 5) == 15

@test("**kwargs")
def test_kwargs():
    def get_values(**kwargs):
        return kwargs
    result = get_values(a=1, b=2)
    assert result["a"] == 1
    assert result["b"] == 2

@test("mixed args")
def test_mixed_args():
    def func(a, b, *args, **kwargs):
        return (a, b, args, kwargs)
    r = func(1, 2, 3, 4, x=5)
    assert r[0] == 1
    assert r[1] == 2
    assert len(r[2]) == 2
    assert r[3]["x"] == 5

# ========== Closures ==========

@test("simple closure")
def test_simple_closure():
    def make_adder(n):
        def adder(x):
            return x + n
        return adder
    add5 = make_adder(5)
    assert add5(10) == 15
    add10 = make_adder(10)
    assert add10(3) == 13

@test("nested closure")
def test_nested_closure():
    def outer(x):
        def middle(y):
            def inner(z):
                return x + y + z
            return inner
        return middle
    assert outer(1)(2)(3) == 6

@test("closure with mutation")
def test_closure_mutation():
    def make_counter():
        count = 0
        def counter():
            nonlocal count
            count = count + 1
            return count
        return counter
    c = make_counter()
    assert c() == 1
    assert c() == 2
    assert c() == 3

# ========== Classes ==========

@test("simple class")
def test_simple_class():
    class Point:
        def __init__(self, x, y):
            self.x = x
            self.y = y
        def add(self, other):
            return Point(self.x + other.x, self.y + other.y)
    p1 = Point(1, 2)
    p2 = Point(3, 4)
    p3 = p1.add(p2)
    assert p3.x == 4
    assert p3.y == 6

@test("class with method")
def test_class_method():
    class Counter:
        def __init__(self):
            self.value = 0
        def increment(self):
            self.value = self.value + 1
            return self.value
    c = Counter()
    assert c.increment() == 1
    assert c.increment() == 2

@test("inheritance")
def test_inheritance():
    class Animal:
        def __init__(self, name):
            self.name = name
        def speak(self):
            return "..."
    class Dog(Animal):
        def speak(self):
            return self.name + " says woof"
    d = Dog("Rex")
    assert d.speak() == "Rex says woof"

@test("class attribute")
def test_class_attribute():
    class MyClass:
        count = 0
        def __init__(self):
            MyClass.count = MyClass.count + 1
    a = MyClass()
    b = MyClass()
    assert MyClass.count == 2

# ========== Control Flow ==========

@test("if/elif/else")
def test_if_elif():
    def classify(x):
        if x < 0:
            return "negative"
        elif x == 0:
            return "zero"
        else:
            return "positive"
    assert classify(-1) == "negative"
    assert classify(0) == "zero"
    assert classify(1) == "positive"

@test("for loop")
def test_for_loop():
    total = 0
    for i in range(5):
        total = total + i
    assert total == 10

@test("while loop")
def test_while_loop():
    x = 0
    while x < 5:
        x = x + 1
    assert x == 5

@test("break statement")
def test_break():
    result = []
    for i in range(10):
        if i == 5:
            break
        result.append(i)
    assert len(result) == 5
    assert result[-1] == 4

@test("continue statement")
def test_continue():
    result = []
    for i in range(10):
        if i % 2 == 0:
            continue
        result.append(i)
    assert len(result) == 5
    assert result == [1, 3, 5, 7, 9]

@test("for-else")
def test_for_else():
    def find(lst, target):
        for x in lst:
            if x == target:
                return True
        else:
            return False
    assert find([1, 2, 3], 2) == True
    assert find([1, 2, 3], 5) == False

# ========== Exceptions ==========

@test("try/except")
def test_try_except():
    def safe_div(a, b):
        try:
            return a / b
        except:
            return None
    assert safe_div(10, 2) == 5.0
    assert safe_div(10, 0) == None

@test("try/except specific")
def test_try_except_specific():
    def get_item(lst, idx):
        try:
            return lst[idx]
        except IndexError:
            return "out of range"
    assert get_item([1, 2, 3], 1) == 2
    assert get_item([1, 2, 3], 10) == "out of range"

@test("try/finally")
def test_try_finally():
    result = []
    def test():
        try:
            result.append("try")
            return "done"
        finally:
            result.append("finally")
    assert test() == "done"
    assert result == ["try", "finally"]

@test("raise exception")
def test_raise():
    def check_positive(x):
        if x < 0:
            raise ValueError("must be positive")
        return True
    assert check_positive(5) == True
    try:
        check_positive(-1)
        assert False, "should have raised"
    except ValueError:
        pass

# ========== Built-in Functions ==========

@test("len()")
def test_len():
    assert len([1, 2, 3]) == 3
    assert len("hello") == 5
    assert len({"a": 1, "b": 2}) == 2

@test("range()")
def test_range():
    r = list(range(5))
    assert r == [0, 1, 2, 3, 4]
    r2 = list(range(2, 6))
    assert r2 == [2, 3, 4, 5]
    r3 = list(range(0, 10, 2))
    assert r3 == [0, 2, 4, 6, 8]

@test("min/max")
def test_min_max():
    assert min([3, 1, 2]) == 1
    assert max([3, 1, 2]) == 3
    assert min(1, 2, 3) == 1
    assert max(1, 2, 3) == 3

@test("sum()")
def test_sum():
    assert sum([1, 2, 3, 4, 5]) == 15

@test("sorted()")
def test_sorted():
    assert sorted([3, 1, 2]) == [1, 2, 3]
    assert sorted([3, 1, 2], reverse=True) == [3, 2, 1]

@test("map/filter")
def test_map_filter():
    doubled = list(map(lambda x: x * 2, [1, 2, 3]))
    assert doubled == [2, 4, 6]
    evens = list(filter(lambda x: x % 2 == 0, [1, 2, 3, 4, 5]))
    assert evens == [2, 4]

@test("zip()")
def test_zip():
    result = list(zip([1, 2, 3], ["a", "b", "c"]))
    assert len(result) == 3
    assert result[0] == (1, "a")

@test("enumerate()")
def test_enumerate():
    result = list(enumerate(["a", "b", "c"]))
    assert result[0] == (0, "a")
    assert result[1] == (1, "b")

# ========== Comprehensions ==========

@test("list comprehension")
def test_list_comprehension():
    squares = [x * x for x in range(5)]
    assert squares == [0, 1, 4, 9, 16]

@test("list comprehension with filter")
def test_list_comp_filter():
    evens = [x for x in range(10) if x % 2 == 0]
    assert evens == [0, 2, 4, 6, 8]

@test("dict comprehension")
def test_dict_comprehension():
    squares = {x: x * x for x in range(5)}
    assert squares[2] == 4
    assert squares[4] == 16

@test("set comprehension")
def test_set_comprehension():
    chars = {c for c in "hello"}
    assert len(chars) == 4  # h, e, l, o

# ========== Decorators ==========

@test("simple decorator")
def test_simple_decorator():
    def double_result(func):
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs) * 2
        return wrapper

    @double_result
    def add(a, b):
        return a + b

    assert add(2, 3) == 10

@test("decorator with args")
def test_decorator_with_args():
    def multiply_by(n):
        def decorator(func):
            def wrapper(*args, **kwargs):
                return func(*args, **kwargs) * n
            return wrapper
        return decorator

    @multiply_by(3)
    def add(a, b):
        return a + b

    assert add(2, 3) == 15

# ========== Advanced Features ==========

@test("lambda")
def test_lambda():
    f = lambda x, y: x + y
    assert f(2, 3) == 5

@test("generator")
def test_generator():
    def gen(n):
        for i in range(n):
            yield i
    result = list(gen(5))
    assert result == [0, 1, 2, 3, 4]

@test("slicing")
def test_slicing():
    lst = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
    assert lst[2:5] == [2, 3, 4]
    assert lst[::2] == [0, 2, 4, 6, 8]
    assert lst[::-1] == [9, 8, 7, 6, 5, 4, 3, 2, 1, 0]

@test("walrus operator")
def test_walrus():
    if (n := 10) > 5:
        assert n == 10

@test("ternary expression")
def test_ternary():
    x = 5
    result = "positive" if x > 0 else "non-positive"
    assert result == "positive"

# ========== String Methods ==========

@test("str.split/join")
def test_str_split_join():
    s = "a,b,c"
    parts = s.split(",")
    assert parts == ["a", "b", "c"]
    joined = "-".join(parts)
    assert joined == "a-b-c"

@test("str.strip")
def test_str_strip():
    s = "  hello  "
    assert s.strip() == "hello"
    assert s.lstrip() == "hello  "
    assert s.rstrip() == "  hello"

@test("str.replace")
def test_str_replace():
    s = "hello world"
    assert s.replace("world", "python") == "hello python"

@test("str.format")
def test_str_format():
    s = "Hello, {}!".format("World")
    assert s == "Hello, World!"
    s2 = "{}, {}!".format("Hello", "World")
    assert s2 == "Hello, World!"

# ========== Dict Methods ==========

@test("dict.keys/values/items")
def test_dict_methods():
    d = {"a": 1, "b": 2}
    keys = list(d.keys())
    values = list(d.values())
    items = list(d.items())
    assert len(keys) == 2
    assert len(values) == 2
    assert len(items) == 2

@test("dict.update")
def test_dict_update():
    d = {"a": 1}
    d.update({"b": 2, "c": 3})
    assert len(d) == 3
    assert d["b"] == 2

@test("dict.pop")
def test_dict_pop():
    d = {"a": 1, "b": 2}
    v = d.pop("a")
    assert v == 1
    assert len(d) == 1
    assert "a" not in d

# ========== List Methods ==========

@test("list.extend")
def test_list_extend():
    a = [1, 2]
    a.extend([3, 4])
    assert a == [1, 2, 3, 4]

@test("list.insert")
def test_list_insert():
    a = [1, 3]
    a.insert(1, 2)
    assert a == [1, 2, 3]

@test("list.remove")
def test_list_remove():
    a = [1, 2, 3, 2]
    a.remove(2)
    assert a == [1, 3, 2]

@test("list.pop")
def test_list_pop():
    a = [1, 2, 3]
    v = a.pop()
    assert v == 3
    assert a == [1, 2]
    v2 = a.pop(0)
    assert v2 == 1
    assert a == [2]

@test("list.sort")
def test_list_sort():
    a = [3, 1, 2]
    a.sort()
    assert a == [1, 2, 3]
    a.sort(reverse=True)
    assert a == [3, 2, 1]

@test("list.reverse")
def test_list_reverse():
    a = [1, 2, 3]
    a.reverse()
    assert a == [3, 2, 1]

# ========== Set Methods ==========

@test("set.union/intersection")
def test_set_union_intersection():
    a = {1, 2, 3}
    b = {2, 3, 4}
    assert a | b == {1, 2, 3, 4}
    assert a & b == {2, 3}

@test("set.add/remove")
def test_set_add_remove():
    s = {1, 2}
    s.add(3)
    assert 3 in s
    s.remove(1)
    assert 1 not in s

# ========== Recursion ==========

@test("recursive function")
def test_recursion():
    def factorial(n):
        if n <= 1:
            return 1
        return n * factorial(n - 1)
    assert factorial(5) == 120
    assert factorial(10) == 3628800

@test("fibonacci")
def test_fibonacci():
    def fib(n):
        if n <= 1:
            return n
        return fib(n - 1) + fib(n - 2)
    assert fib(10) == 55

# ========== Main ==========

if __name__ == "__main__":
    success = run_tests()
    sys.exit(0 if success else 1)
