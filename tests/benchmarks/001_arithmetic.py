print("=== 算术运算基准测试 ===")

def fib(n):
    if n &lt;= 1:
        return n
    return fib(n-1) + fib(n-2)

def factorial(n):
    result = 1
    for i in range(1, n+1):
        result *= i
    return result

result_fib = fib(20)
result_fact = factorial(10)

print("斐波那契(20) =", result_fib)
print("阶乘(10) =", result_fact)

# 简单的循环运算
sum_result = 0
for i in range(10000):
    sum_result += i

print("1-9999的和 =", sum_result)
