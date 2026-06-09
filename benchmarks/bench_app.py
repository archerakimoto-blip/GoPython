# 综合应用性能测试
# 测试接近真实场景的综合计算性能

import time
import math

# 快速排序
def quicksort(arr):
    if len(arr) <= 1:
        return arr
    pivot = arr[len(arr) // 2]
    left = []
    middle = []
    right = []
    for x in arr:
        if x < pivot:
            left.append(x)
        elif x == pivot:
            middle.append(x)
        else:
            right.append(x)
    return quicksort(left) + middle + quicksort(right)

start = time.time()
data = [1000 - i for i in range(1000)]
result = quicksort(data)
elapsed_quicksort = time.time() - start

# 矩阵乘法 (纯 Python)
def matrix_mul(a, b):
    n = len(a)
    c = [[0] * n for _ in range(n)]
    i = 0
    while i < n:
        j = 0
        while j < n:
            s = 0
            k = 0
            while k < n:
                s = s + a[i][k] * b[k][j]
                k = k + 1
            c[i][j] = s
            j = j + 1
        i = i + 1
    return c

n = 50
a = [[i + j for j in range(n)] for i in range(n)]
b = [[i * j for j in range(n)] for i in range(n)]
start = time.time()
c = matrix_mul(a, b)
elapsed_matrix = time.time() - start

# N-Queens 问题
def solve_nqueens(n):
    count = 0
    def backtrack(row, cols, diag1, diag2):
        nonlocal count
        if row == n:
            count = count + 1
            return
        col = 0
        while col < n:
            if col in cols or (row - col) in diag1 or (row + col) in diag2:
                col = col + 1
                continue
            cols.append(col)
            diag1.append(row - col)
            diag2.append(row + col)
            backtrack(row + 1, cols, diag1, diag2)
            cols.pop()
            diag1.pop()
            diag2.pop()
            col = col + 1
    backtrack(0, [], [], [])
    return count

start = time.time()
x = solve_nqueens(10)
elapsed_nqueens = time.time() - start

# 斐波那契 (迭代)
def fib_iter(n):
    if n <= 1:
        return n
    a = 0
    b = 1
    i = 2
    while i <= n:
        a, b = b, a + b
        i = i + 1
    return b

start = time.time()
x = fib_iter(1000000)
elapsed_fib_iter = time.time() - start

# 字符串处理 - 词频统计
text = "the quick brown fox jumps over the lazy dog " * 100
start = time.time()
words = text.split()
freq = {}
for w in words:
    if w in freq:
        freq[w] = freq[w] + 1
    else:
        freq[w] = 1
elapsed_wordcount = time.time() - start

# 曼德博集合计算 (简化版)
def mandelbrot(max_iter):
    count = 0
    y = -2
    while y <= 2:
        x = -2
        while x <= 2:
            zr = 0.0
            zi = 0.0
            cr = x
            ci = y
            i = 0
            while i < max_iter:
                zr2 = zr * zr
                zi2 = zi * zi
                if zr2 + zi2 > 4:
                    break
                zi = 2 * zr * zi + ci
                zr = zr2 - zi2 + cr
                i = i + 1
            if i == max_iter:
                count = count + 1
            x = x + 0.1
        y = y + 0.1
    return count

start = time.time()
x = mandelbrot(50)
elapsed_mandelbrot = time.time() - start

print("BENCHMARK_RESULT|app_quicksort_1k|{:.6f}".format(elapsed_quicksort))
print("BENCHMARK_RESULT|app_matrix_mul_50|{:.6f}".format(elapsed_matrix))
print("BENCHMARK_RESULT|app_nqueens_10|{:.6f}".format(elapsed_nqueens))
print("BENCHMARK_RESULT|app_fib_iter_1m|{:.6f}".format(elapsed_fib_iter))
print("BENCHMARK_RESULT|app_wordcount|{:.6f}".format(elapsed_wordcount))
print("BENCHMARK_RESULT|app_mandelbrot|{:.6f}".format(elapsed_mandelbrot))
