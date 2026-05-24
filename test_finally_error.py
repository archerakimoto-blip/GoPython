try:
    print("Try")
    x = 1 / 0
except:
    print("Except")
finally:
    print("Finally")
    y = 2 / 0
    print("After finally error")
print("Done")