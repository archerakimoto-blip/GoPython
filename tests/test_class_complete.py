# 类系统完整测试套件

print("=" * 70)
print("GoPy 类系统测试")
print("=" * 70)

# Test 1: 简单类定义
print("\n[Test 1] 简单类定义")
class Dog:
    __init__(self, name):
        self.name = name
    
    bark(self):
        print("Woof!")

print("Dog 类已定义")

# Test 2: 实例化
print("\n[Test 2] 实例化")
d = Dog("Buddy")
print("实例已创建")

# Test 3: 属性访问
print("\n[Test 3] 属性访问")
print(d.name)

# Test 4: 属性赋值
print("\n[Test 4] 属性赋值")
d.name = "Max"
print(d.name)

# Test 5: 方法调用
print("\n[Test 5] 方法调用")
d.bark()

# Test 6: 多实例
print("\n[Test 6] 多实例")
d2 = Dog("Charlie")
print(d.name)
print(d2.name)

# Test 7: 类属性
print("\n[Test 7] 简单 Point 类")
class Point:
    __init__(self, x, y):
        self.x = x
        self.y = y
    
    get_x(self):
        return self.x
    
    get_y(self):
        return self.y
    
    set_x(self, x):
        self.x = x
    
    set_y(self, y):
        self.y = y

p = Point(1, 2)
print(p.get_x())
print(p.get_y())
p.set_x(10)
p.set_y(20)
print(p.get_x())
print(p.get_y())

print("\n" + "=" * 70)
print("所有测试完成！")
print("=" * 70)
