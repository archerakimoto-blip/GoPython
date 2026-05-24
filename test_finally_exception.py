try:
    print("Try")
    x = 1 / 0
    print("After division")
except:
    print("Except")
finally:
    print("Finally")
print("Done")