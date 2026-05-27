import time

def test_basic_concurrent():
    print("Testing basic concurrency...")
    
    def worker(id, result_list):
        for i in range(1000):
            result_list.append(i * id)
        return id * 10
    
    results = []
    threads = []
    
    for i in range(5):
        t = threading.Thread(target=worker, args=(i, results))
        threads.append(t)
        t.start()
    
    for t in threads:
        t.join()
    
    print(f"Total results: {len(results)}")
    print(f"First few results: {results[:10]}")
    print("Basic concurrency test passed!")

def test_channel():
    print("\nTesting channels...")
    
    ch = make(10)
    
    def producer(ch):
        for i in range(20):
            ch.send(i)
            time.sleep(0.001)
        ch.close()
    
    def consumer(ch, results):
        while True:
            try:
                val = ch.receive()
                results.append(val)
            except:
                break
    
    results = []
    t1 = threading.Thread(target=producer, args=(ch,))
    t2 = threading.Thread(target=consumer, args=(ch, results))
    
    t1.start()
    t2.start()
    
    t1.join()
    t2.join()
    
    print(f"Received {len(results)} items")
    print(f"Results: {results}")
    print("Channel test passed!")

def test_atomic():
    print("\nTesting atomic operations...")
    
    counter = AtomicInteger(0)
    
    def incrementer(counter, iterations):
        for _ in range(iterations):
            counter.increment()
    
    threads = []
    for _ in range(10):
        t = threading.Thread(target=incrementer, args=(counter, 10000))
        threads.append(t)
        t.start()
    
    for t in threads:
        t.join()
    
    expected = 10 * 10000
    actual = counter.get()
    
    print(f"Expected: {expected}, Actual: {actual}")
    assert actual == expected, f"Atomic test failed: {actual} != {expected}"
    print("Atomic operations test passed!")

if __name__ == "__main__":
    import threading
    from concurrent import make as make_channel, AtomicInteger
    
    test_basic_concurrent()
    test_channel()
    test_atomic()
    print("\n=== All concurrency tests passed! ===")