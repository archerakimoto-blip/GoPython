# GoPy 开发计划

## 设计理念

GoPy 采用**脱糖优先**（Desugar-First）的架构设计。核心原则是：**尽可能在脱糖层处理高级语法特性，将编译器和虚拟机保持为最小内核**。

这样做的好处：
1. **内核简洁** — 编译器和虚拟机只需处理基础语义，无需感知装饰器、slots 等高级特性
2. **可维护性** — 高级特性的变更集中在脱糖层，不污染内核代码
3. **可测试性** — 脱糖输出是纯 AST，可以独立验证转换正确性
4. **可扩展性** — 新增语法特性只需在脱糖层添加转换规则

---

## 已完成

### v0.3 — 装饰器脱糖重构（当前版本）

将 `@property`、`@classmethod`、`@staticmethod` 和 `__slots__` 的处理从内核迁移至脱糖层。

#### 变更详情

| 模块 | 变更 |
|------|------|
| `pkg/desugar/desugar.go` | 新增 `desugarClassStatement` 函数，处理类级别的装饰器和 `__slots__` 脱糖 |
| `pkg/objects/object.go` | 新增 `BoundMethod` 类型和 `BOUNDMETHOD_OBJ` 类型常量；移除 `Property`、`ClassMethod`、`StaticMethod` 类型和 `Class.Slots` 字段 |
| `pkg/compiler/compiler.go` | 新增内置函数 `__get_class__`、`__bind_method__`、`__set_field__`、`__del_field__`；移除装饰器编译逻辑 |
| `pkg/vm/vm.go` | `executeCall` 新增 `BoundMethod` 展开（将 `bm.Fn` + `bm.Self` + args 推入栈）；移除 `Property`/`ClassMethod`/`StaticMethod` 运行时处理 |

#### 脱糖转换规则

**`@property`**

```python
class C:
    @property
    def x(self):
        return self._x
```

脱糖为：

```python
class C:
    def _desugar_prop_get_x(self):
        return self._x

    def __getattr__(self, name):
        if name == 'x':
            return self._desugar_prop_get_x()
```

**`@property` + `@x.setter`**

```python
class C:
    @property
    def x(self):
        return self._x

    @x.setter
    def x(self, value):
        self._x = value
```

脱糖为：

```python
class C:
    def _desugar_prop_get_x(self):
        return self._x

    def _desugar_prop_set_x(self, value):
        self._x = value

    def __getattr__(self, name):
        if name == 'x':
            return self._desugar_prop_get_x()

    def __setattr__(self, name, value):
        if name == 'x':
            self._desugar_prop_set_x(value)
            return
        __set_field__(self, name, value)
```

**`@classmethod`**

```python
class C:
    @classmethod
    def foo(cls, arg):
        return arg
```

脱糖为：

```python
class C:
    def _desugar_cm_foo(cls, arg):
        return arg

    def __getattr__(self, name):
        if name == 'foo':
            return __bind_method__(__get_class__(self)._desugar_cm_foo, __get_class__(self))
```

**`@staticmethod`**

```python
class C:
    @staticmethod
    def bar(arg):
        return arg
```

脱糖为：

```python
class C:
    def _desugar_sm_bar(arg):
        return arg

    def __getattr__(self, name):
        if name == 'bar':
            return __get_class__(self)._desugar_sm_bar
```

**`__slots__`**

```python
class C:
    __slots__ = ['x', 'y']
```

脱糖为：

```python
class C:
    def __setattr__(self, name, value):
        if name == 'x':
            __set_field__(self, name, value)
            return
        if name == 'y':
            __set_field__(self, name, value)
            return
        raise AttributeError("'C' object has no attribute '" + name + "'")
```

---

## 进行中

### 脱糖层增强

- [ ] `@property` 的 `@x.deleter` 端到端测试
- [ ] `@classmethod` / `@staticmethod` 端到端测试
- [ ] `__slots__` 端到端测试
- [ ] 装饰器与继承的交互测试
- [ ] 用户自定义 `__getattr__` / `__setattr__` 与脱糖生成代码的合并测试

### 词法分析器 / 解析器

- [ ] 类体中 `@property` / `@classmethod` / `@staticmethod` 装饰器解析的健壮性改进
- [ ] INDENT/DEDENT 边界情况处理

---

## 计划中

### v0.4 — 脱糖层扩展

- [ ] `@dataclass` 装饰器脱糖 — 自动生成 `__init__`、`__repr__`、`__eq__` 方法
- [ ] `@abstractmethod` 装饰器脱糖 — 在 `__getattr__` 中添加抽象检查
- [ ] `@functools.lru_cache` 装饰器脱糖 — 生成缓存包装方法
- [ ] 多重继承 MRO（方法解析顺序）在脱糖层实现
- [ ] `super()` 在脱糖层展开为直接父类方法调用

### v0.5 — 性能优化

- [ ] 脱糖层生成的 `__getattr__` / `__setattr__` 优化：使用哈希表替代 if-else 链
- [ ] BoundMethod 缓存：避免重复创建
- [ ] `__slots__` 脱糖生成的 `__setattr__` 编译为跳转表而非线性搜索

### v0.6 — 类型系统

- [ ] 类型注解脱糖 — 将类型提示转换为运行时检查或静态分析数据
- [ ] `typing` 模块基础支持

---

## 架构约束

1. **脱糖层不修改内核** — 脱糖层只能生成由现有内核支持的 AST 节点和内置函数调用
2. **内置函数最小化** — 新增内置函数需有充分理由，优先使用现有指令组合
3. **向后兼容** — 脱糖转换不应改变用户可见的语义行为
4. **可调试性** — 脱糖生成的混淆方法名（`_desugar_*`）应保持可追溯性
