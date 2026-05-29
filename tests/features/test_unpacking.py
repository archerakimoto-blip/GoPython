def test_list_unpacking():
    a = [1, 2, 3]
    b = [*a, 4, 5]
    print(b)
    assert b == [1, 2, 3, 4, 5]

def test_dict_unpacking():
    d1 = {"a": 1, "b": 2}
    d2 = {"b": 3, "c": 4}
    combined = {**d1, **d2, "d": 5}
    print(combined)
    assert combined == {"a": 1, "b": 3, "c": 4, "d": 5}

def test_mixed_unpacking():
    # 混合解包
    list1 = [1, 2]
    list2 = [3, 4]
    result = [*list1, 5, *list2, 6]
    print(result)
    assert result == [1, 2, 5, 3, 4, 6]
    
    # 字典混合解包
    d1 = {"x": 1}
    d2 = {"y": 2}
    result_dict = {**d1, "z": 3, **d2}
    print(result_dict)
    assert result_dict == {"x": 1, "z": 3, "y": 2}

def run_tests():
    print("Testing list unpacking...")
    test_list_unpacking()
    print("✓ List unpacking test passed")
    
    print("\nTesting dict unpacking...")
    test_dict_unpacking()
    print("✓ Dict unpacking test passed")
    
    print("\nTesting mixed unpacking...")
    test_mixed_unpacking()
    print("✓ Mixed unpacking test passed")
    
    print("\n✅ All unpacking tests passed!")

if __name__ == "__main__":
    run_tests()
