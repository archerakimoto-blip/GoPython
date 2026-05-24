# GoPy 类系统测试

print("=" * 60)
print("测试类定义和实例化")
print("=" * 60)

class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y
    
    def print(self):
        print("Point:")
        print(self.x)
        print(self.y)

print("\n类已定义")

p = Point(1, 2)
print("\n实例已创建:")
print(p)

print("\n访问属性:")
print(p.x)
print(p.y)

print("\n修改属性:")
p.x = 10
p.y = 20
print(p.x)
print(p.y)

print("\n调用方法:")
p.print()

print("\n" + "=" * 60)
print("所有测试完成！")
print("=" * 60)
