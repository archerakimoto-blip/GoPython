"""
Matrix and Array Operations Benchmark
Tests 2D array and matrix operations
"""

def create_matrix(size):
    matrix = []
    i = 0
    while i < size:
        row = []
        j = 0
        while j < size:
            row.append(i * size + j)
            j = j + 1
        matrix.append(row)
        i = i + 1
    return matrix

def matrix_addition(a, b):
    result = []
    i = 0
    while i < len(a):
        row = []
        j = 0
        while j < len(a[0]):
            row.append(a[i][j] + b[i][j])
            j = j + 1
        result.append(row)
        i = i + 1
    return result

def matrix_scalar_multiply(matrix, scalar):
    result = []
    i = 0
    while i < len(matrix):
        row = []
        j = 0
        while j < len(matrix[0]):
            row.append(matrix[i][j] * scalar)
            j = j + 1
        result.append(row)
        i = i + 1
    return result

def matrix_sum(matrix):
    total = 0
    i = 0
    while i < len(matrix):
        j = 0
        while j < len(matrix[0]):
            total = total + matrix[i][j]
            j = j + 1
        i = i + 1
    return total

def test_flat_list_operations(size):
    lst = []
    i = 0
    while i < size * size:
        lst.append(i)
        i = i + 1
    
    total = 0
    i = 0
    while i < size * size:
        total = total + lst[i] * 2
        i = i + 1
    return total

# Run benchmarks
print("=== 矩阵操作基准测试 ===")
print()

print("1. 矩阵创建 (100x100):")
m1 = create_matrix(100)
m2 = create_matrix(100)
print("  完成")

print()
print("2. 矩阵加法 (100x100):")
_ = matrix_addition(m1, m2)
print("  完成")

print()
print("3. 矩阵标量乘法 (100x100):")
_ = matrix_scalar_multiply(m1, 2)
print("  完成")

print()
print("4. 矩阵求和 (100x100):")
sum_result = matrix_sum(m1)
print("  结果:", sum_result)

print()
print("5. 扁平列表操作 (10000元素):")
_ = test_flat_list_operations(100)
print("  完成")

print()
print("=== 矩阵测试完成 ===")
