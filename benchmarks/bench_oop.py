# 面向对象性能测试
# 测试类定义、实例化、方法调用、继承等性能

import time

N = 50000

# 简单类实例化
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y

start = time.time()
i = 0
while i < N:
    p = Point(i, i + 1)
    i = i + 1
elapsed_instantiation = time.time() - start

# 方法调用
class Counter:
    def __init__(self):
        self.count = 0

    def increment(self):
        self.count = self.count + 1
        return self.count

    def get_count(self):
        return self.count

start = time.time()
c = Counter()
i = 0
while i < N:
    c.increment()
    i = i + 1
elapsed_method_call = time.time() - start

# 属性访问
start = time.time()
p = Point(1, 2)
i = 0
x = 0
while i < N:
    x = p.x
    i = i + 1
elapsed_attr_access = time.time() - start

# 属性赋值
start = time.time()
p = Point(0, 0)
i = 0
while i < N:
    p.x = i
    i = i + 1
elapsed_attr_assign = time.time() - start

# 继承
class Animal:
    def __init__(self, name):
        self.name = name

    def speak(self):
        return "..."

class Dog(Animal):
    def speak(self):
        return "Woof"

start = time.time()
i = 0
x = ""
while i < N:
    d = Dog("Rex")
    x = d.speak()
    i = i + 1
elapsed_inheritance = time.time() - start

# 多层继承
class Base:
    def method(self):
        return 1

class Mid(Base):
    def method(self):
        return 2

class Leaf(Mid):
    def method(self):
        return 3

start = time.time()
obj = Leaf()
i = 0
x = 0
while i < N:
    x = obj.method()
    i = i + 1
elapsed_multi_inherit = time.time() - start

print("BENCHMARK_RESULT|oop_instantiation|{:.6f}".format(elapsed_instantiation))
print("BENCHMARK_RESULT|oop_method_call|{:.6f}".format(elapsed_method_call))
print("BENCHMARK_RESULT|oop_attr_access|{:.6f}".format(elapsed_attr_access))
print("BENCHMARK_RESULT|oop_attr_assign|{:.6f}".format(elapsed_attr_assign))
print("BENCHMARK_RESULT|oop_inheritance|{:.6f}".format(elapsed_inheritance))
print("BENCHMARK_RESULT|oop_multi_inherit|{:.6f}".format(elapsed_multi_inherit))
