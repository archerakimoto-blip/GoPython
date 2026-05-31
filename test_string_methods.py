let s = "hello, world!"
print("Original:", s)
print("Upper:", s.upper())
print("Lower:", s.lower())
print("Capitalize:", s.capitalize())
print("Title:", s.title())
print("Swapcase:", s.swapcase())
print("Replace 'l' with 'x':", s.replace("l", "x"))
print("Starts with 'hello':", s.startswith("hello"))
print("Ends with '!':", s.endswith("!"))
print("Find 'world':", s.find("world"))
print("Count 'l':", s.count("l"))
print("Split:", s.split())
print("Strip:", s.strip())

let join_sep = "-"
let words = ["hello", "world", "test"]
print("Join:", join_sep.join(words))

let is_str = "hello123"
print("Is alpha:", is_str.isalpha())
print("Is digit:", is_str.isdigit())
print("Is alnum:", is_str.isalnum())
print("Is space:", "   ".isspace())
print("Is upper:", "HELLO".isupper())
print("Is lower:", "hello".islower())

# Test list methods
let lst = [1, 2, 3, 4, 5]
print("\nList before append:", lst)
lst.append(6)
print("List after append:", lst)
lst.extend([7, 8])
print("List after extend:", lst)
lst.pop()
print("List after pop:", lst)
lst.pop(0)
print("List after pop index 0:", lst)
lst.insert(0, 1)
print("List after insert:", lst)
lst.remove(2)
print("List after remove:", lst)
print("Index of 3:", lst.index(3))
print("Count of 4:", lst.count(4))
lst.reverse()
print("List after reverse:", lst)

# Test dict methods
let d = {"a": 1, "b": 2, "c": 3}
print("\nDict keys:", d.keys())
print("Dict values:", d.values())
print("Dict items:", d.items())
print("Get 'b':", d.get("b"))
print("Get 'd' default 0:", d.get("d", 0))
d.setdefault("d", 4)
print("Dict after setdefault:", d)
d.update({"e":5, "f":6})
print("Dict after update:", d)
d.pop("a")
print("Dict after pop a:", d)
