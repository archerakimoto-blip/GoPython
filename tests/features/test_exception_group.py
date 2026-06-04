# Test ExceptionGroup basic creation
eg = ExceptionGroup("group", [TypeError("a"), ValueError("b")])
print(eg)

# Test except* with ExceptionGroup
try:
    raise ExceptionGroup("eg", [TypeError("err1"), ValueError("err2")])
except* TypeError as e:
    print("caught TypeError group:", e)

# Test except* with multiple handlers
try:
    raise ExceptionGroup("eg", [TypeError("err1"), ValueError("err2")])
except* TypeError as e:
    print("caught TypeError:", e)
except* ValueError as e:
    print("caught ValueError:", e)

# Test except* with single exception wrapped
try:
    raise ExceptionGroup("eg", [TypeError("err1")])
except* TypeError as e:
    print("caught single TypeError:", e)
