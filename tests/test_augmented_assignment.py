# 测试增强赋值
a = 10
a += 5
print("a += 5:", a)

b = 20
b -= 3
print("b -= 3:", b)

c = 5
c *= 4
print("c *= 4:", c)

d = 20
d /= 2
print("d /= 2:", d)

# 位运算增强赋值 (使用十进制数字)
x = 10  # 0b1010
x |= 5  # 0b0101
print("x |= 5:", x)

y = 15  # 0b1111
y &= 10 # 0b1010
print("y &= 10:", y)

z = 10  # 0b1010
z ^= 12 # 0b1100
print("z ^= 12:", z)

shift = 1
shift <<= 3
print("shift <<= 3:", shift)

shift2 = 16
shift2 >>= 2
print("shift2 >>= 2:", shift2)