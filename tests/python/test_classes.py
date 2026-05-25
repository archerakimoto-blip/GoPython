print("=== Class and Object System Tests ===")

class Person {
    def greet(self) {
        return "Hello!"
    }
    
    def get_name(self) {
        return "Person"
    }
}

p = Person()
print("Creating Person instance:")
print(p)

print("\nCalling p.greet():")
print(p.greet())

print("\nCalling p.get_name():")
print(p.get_name())

print("\nSetting attribute:")
p.name = "Alice"
p.age = 30

print("p.name:")
print(p.name)
print("p.age:")
print(p.age)

print("\n=== Class with Constructor ===")

class Calculator {
    def __init__(self, value) {
        self.result = value
    }
    
    def add(self, x) {
        self.result = self.result + x
        return self.result
    }
    
    def get_result(self) {
        return self.result
    }
}

calc = Calculator(10)
print("Created calculator with initial value 10:")
print(calc.get_result())

print("\ncalc.add(5):")
print(calc.add(5))

print("\ncalc.add(3):")
print(calc.add(3))

print("\n=== Method Chaining ===")

class StringBuilder {
    def __init__(self) {
        self.str = ""
    }
    
    def append(self, text) {
        self.str = self.str + text
        return self
    }
    
    def to_string(self) {
        return self.str
    }
}

sb = StringBuilder()
result = sb.append("Hello").append(" ").append("World").to_string()
print("StringBuilder result:")
print(result)

print("\n=== Nested Objects ===")

class Container {
    def __init__(self) {
        self.items = []
    }
    
    def add_item(self, item) {
        self.items.append(item)
    }
    
    def get_items(self) {
        return self.items
    }
    
    def size(self) {
        return len(self.items)
    }
}

container = Container()
container.add_item("apple")
container.add_item("banana")
container.add_item("orange")

print("Container items:")
items = container.get_items()
print(items)

print("\nContainer size:")
print(container.size())

print("\n=== GetAttr on Module ===")

print("Accessing math.sqrt through module:")
sqrt_val = math.sqrt
print(sqrt_val)

print("\nCalling math.sqrt(25) directly:")
print(math.sqrt(25))

print("\nAll class and object tests passed!")
