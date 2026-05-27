
# Test <= operator
a = 5
if a <= 10:
    print("a <= 10: yes")

if a <= 5:
    print("a <=5: yes")

# Test >= operator
b = 20
if b >= 15:
    print("b >=15: yes")

if b >= 20:
    print("b >=20: yes")

# Test if block followed by another statement
def test(x):
    if x <= 0:
        return 99
    return x * 2

print("test(-3):", test(-3))
print("test(4):", test(4))
