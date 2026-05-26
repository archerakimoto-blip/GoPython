
# 测试新功能
print("=== 测试改进的 min/max 函数 ===");
print("min(3, 1, 4, 1, 5) =", min(3, 1, 4, 1, 5));
print("min(3.14, 1.59, 2.65) =", min(3.14, 1.59, 2.65));
print("min(5, 3.14, 7) =", min(5, 3.14, 7));
print();

print("max(3, 1, 4, 1, 5) =", max(3, 1, 4, 1, 5));
print("max(3.14, 1.59, 2.65) =", max(3.14, 1.59, 2.65));
print("max(5, 3.14, 7) =", max(5, 3.14, 7));
print();

print("=== 测试改进的 sum 函数 ===");
print("sum([1, 2, 3, 4, 5]) =", sum([1, 2, 3, 4, 5]));
print("sum([1.5, 2.5, 3.5]) =", sum([1.5, 2.5, 3.5]));
print("sum([1, 2.5, 3, 4.5]) =", sum([1, 2.5, 3, 4.5]));
print();

print("=== 测试 zip 函数 ===");
a = [1, 2, 3];
b = ["one", "two", "three"];
c = [True, False, True];
zipped = zip(a, b, c);
print("zip([1,2,3], ['one','two','three'], [True,False,True]) =", zipped);
print();

print("=== 测试增强的 math 模块 ===");
import math;
print("math.pi =", math.pi);
print("math.e =", math.e);
print();

print("三角函数:");
print("math.sin(math.pi/2) =", math.sin(math.pi/2));
print("math.cos(0) =", math.cos(0));
print("math.tan(math.pi/4) =", math.tan(math.pi/4));
print("math.asin(1) =", math.asin(1));
print("math.acos(0) =", math.acos(0));
print("math.atan(1) =", math.atan(1));
print();

print("取整函数:");
print("math.floor(3.9) =", math.floor(3.9));
print("math.ceil(3.1) =", math.ceil(3.1));
print("math.trunc(3.9) =", math.trunc(3.9));
print("math.trunc(-3.9) =", math.trunc(-3.9));
print();

print("指数和对数:");
print("math.exp(1) =", math.exp(1));
print("math.log(math.e) =", math.log(math.e));
print("math.log10(100) =", math.log10(100));
print("math.log2(8) =", math.log2(8));
print("math.sqrt(16) =", math.sqrt(16));
print("math.pow(2, 3) =", math.pow(2, 3));
print("math.hypot(3, 4) =", math.hypot(3, 4));
print();

print("角度和弧度转换:");
print("math.degrees(math.pi) =", math.degrees(math.pi));
print("math.radians(180) =", math.radians(180));
print();

print("所有测试通过!");
