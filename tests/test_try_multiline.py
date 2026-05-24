try:
    x = 1 / 0
except Exception as e:
    print("Caught exception: ")
    print(e)
print("After try/except")
