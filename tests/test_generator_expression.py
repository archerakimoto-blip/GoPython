# 测试生成器表达式
gen = (x for x in [1, 2, 3, 4, 5])
print("Generator type:", type(gen))
print("Generator elements:", list(gen))

# 带条件的生成器表达式
gen2 = (x * 2 for x in range(5) if x % 2 == 0)
print("Filtered generator:", list(gen2))

# 带运算的生成器表达式
gen3 = (x ** 2 for x in range(1, 4))
print("Squared generator:", list(gen3))

# 用于函数参数
print("Sum from generator:", sum(x for x in range(1, 6)))