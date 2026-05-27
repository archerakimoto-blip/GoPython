"""测试 CPython 互操作模块"""
print("=== 测试 CPython 互操作 ===")
print()

# 测试 cpython 模块
print("测试 cpython 模块:")

# 执行简单表达式
import cpython
result = cpython.eval("2 + 3 * 4")
print("  cpython.eval('2 + 3 * 4'):", result)

# 执行代码块
cpython.exec("x = 10; y = 20; z = x + y")
result = cpython.eval("z")
print("  cpython.exec('x = 10; y = 20; z = x + y')")
print("  cpython.eval('z'):", result)

# 定义函数并调用
cpython.exec("""
def greet(name):
    return f'Hello, {name}!'
""")
result = cpython.eval("greet('World')")
print("  cpython.eval(\"greet('World')\"):", result)

print()
print("=== CPython 互操作测试完成 ===")