import math
import os
import json

print("=== Testing os module ===")
print("os.sep:", os.sep)
print("os.getcwd():", os.getcwd())

print("\n=== Testing json module ===")
data_dict = {
    "name": "Bob",
    "age": 25,
    "is_student": True,
    "courses": ["Math", "Science", "English"],
    "score": 95.5
}
print("Original dict:", data_dict)

json_str = json.dumps(data_dict)
print("Serialized JSON:", json_str)

parsed_data = json.loads(json_str)
print("Parsed data:", parsed_data)
