def test_with():
    print("Testing with statement")
    with open("test.txt", "r") as f:
        print("Inside with block")
        print(f)
    print("After with block")
test_with()
