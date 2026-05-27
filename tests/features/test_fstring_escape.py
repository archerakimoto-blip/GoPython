name = "Bob";

# 测试转义花括号
test1 = f"Hello {{name}} will be replaced: {name}";
print(test1);

test2 = f"Double braces become single: {{}} and {name}";
print(test2);
