# try/except/finally 异常处理测试

try:
    x = 1 / 0
    print("should not reach here")
except Exception as e:
    print("Caught exception: ")
    print(e)

print("After try/except")
