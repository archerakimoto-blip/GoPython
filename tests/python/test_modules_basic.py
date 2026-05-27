import math
import sys
import os
import json

print("=== All modules imported successfully ===")
print("=== sys module ===")
print("sys.version:", sys.version)
print("sys.platform:", sys.platform)
print("sys.version_info:", sys.version_info)
print("sys.argv:", sys.argv)
print("sys.path:", sys.path)
print("sys.getsizeof(100):", sys.getsizeof(100))

print("\n=== os module ===")
print("os.sep:", os.sep)
print("os.getcwd():", os.getcwd())

print("\n=== json module ===")
test_str = json.dumps("Hello, GoPython!")
print("json.dumps('Hello'):", test_str)
parsed_val = json.loads(test_str)
print("json.loads():", parsed_val)
