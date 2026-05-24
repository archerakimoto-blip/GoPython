# GoPy 类系统实现方案

## 当前状态

已完成的准备工作：

### 1. ✅ AST 定义 ([ast.go](file:///workspace/pkg/ast/ast.go))
添加了以下新节点：
- `ClassStatement` - 类定义语句
- `ClassInstantiation` - 类实例化表达式
- `MemberAccess` - 属性访问表达式 (`obj.attr`)
- `MethodCall` - 方法调用表达式 (`obj.method()`)

### 2. ✅ 对象系统 ([object.go](file:///workspace/pkg/objects/object.go))
添加了新的对象类型：
- `CLASS_OBJ` 和 `INSTANCE_OBJ` 常量
- `Class` 类型 - 包含 Name, Methods, Fields, SuperClass
- `Instance` 类型 - 包含 Class, Fields, InitMethod
- 实例方法：
  - `GetAttr(name)` - 获取属性
  - `SetAttr(name, value)` - 设置属性

### 3. ✅ 关键字 ([lexer.go](file:///workspace/pkg/lexer/lexer.go))
- `CLASS` 关键字已存在（第 56 行）
- 关键字映射已添加（第 87 行）

## 待完成的工作

### 1. Parser ([parser.go](file:///workspace/pkg/parser/parser.go))
需要添加：
- `parseClassStatement()` - 解析 `class ClassName:` 语句
- 处理类体中的方法定义
- 处理 `obj.attr` 语法
- 处理 `obj.method()` 语法

### 2. Compiler ([compiler.go](file:///workspace/pkg/compiler/compiler.go))
需要添加：
- `ClassLiteral` - 编译类定义为特殊对象
- 类实例化编译
- 属性访问编译
- 方法调用编译

### 3. VM ([vm.go](file:///workspace/pkg/vm/vm.go))
需要添加：
- 新的操作码：
  - `OpCreateClass` - 创建类对象
  - `OpInstantiateClass` - 实例化类
  - `OpGetAttribute` - 获取属性
  - `OpSetAttribute` - 设置属性
- 属性和方法查找逻辑
- `__init__` 方法自动调用

## 实现策略

### 简化版本（当前目标）

只实现基本功能，不支持：
- 继承（inheritance）
- 多重继承
- 类变量
- 静态方法
- 类方法

### 语法支持

```python
class ClassName:
    def __init__(self, param):
        self.attr = param
    
    def method(self):
        return self.attr

obj = ClassName(value)
obj.attr = new_value
print(obj.attr)
obj.method()
```

## 预期实现时间

1. Parser 修改：1-2 小时
2. Compiler 修改：2-3 小时
3. VM 修改：2-3 小时
4. 测试：1-2 小时

总计：约 6-10 小时

## 测试用例

已创建测试文件：
- [test_class_basic.py](file:///workspace/tests/test_class_basic.py) - 基本类功能测试

## 下一步

1. 修改 parser.go 添加类定义解析
2. 添加方法调用的前缀和infix解析器
3. 在 compiler 中添加类编译逻辑
4. 在 VM 中添加相应的操作码处理
5. 集成测试和调试
