# try/except/finally 完整测试

try:
    print("Try block")
    x = 10 / 0
    print("After division")
except Exception as e:
    print("Except block")
    print("Caught: ")
    print(e)
finally:
    print("Finally block")

print("After everything")
