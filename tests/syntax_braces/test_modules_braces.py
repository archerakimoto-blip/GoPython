import sys
import os
import json

print("=== Testing sys module ===")
print("sys.version:", sys.version)
print("sys.platform:", sys.platform)
print("sys.version_info:", sys.version_info)
print("sys.argv:", sys.argv)
print("sys.path:", sys.path)
print("sys.getsizeof(42):", sys.getsizeof(42))

print("\n=== Testing os module ===")
print("os.sep:", os.sep)
print("os.getcwd():", os.getcwd())

print("\n=== Testing json module ===")
data = {"name": "Alice", "age": 30, "city": "New York"}
print("Original data:", data)
json_str = json.dumps(data)
print("JSON string:", json_str)
loaded_data = json.loads(json_str)
print("Loaded data:", loaded_data)
