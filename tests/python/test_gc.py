#!/usr/bin/env python
"""测试垃圾回收功能"""

import gc

print("=== GoPython 垃圾回收测试 ===\n")

# 测试 1: 启用/禁用 GC
print("1. 测试 GC 启用/禁用")
gc.disable()
print("   GC 已禁用")
gc.enable()
print("   GC 已启用\n")

# 测试 2: 获取 GC 统计信息
print("2. 测试 GC 统计信息")
stats = gc.get_stats()
print("   GC 统计信息:")
print(f"   - 集合总数: {stats['collection_count']}")
print(f"   - 已分配字节数: {stats['allocated_bytes']}")
print(f"   - 已释放字节数: {stats['freed_bytes']}")
print(f"   - 存活对象数: {stats['live_objects']}\n")

# 测试 3: 打印 GC 统计信息
print("3. 打印详细 GC 统计信息:")
gc.print_stats()
print()

# 测试 4: 手动触发 GC
print("4. 手动触发 GC...")
gc.collect()
print("   GC 完成！")

# 再次获取统计信息
stats_after = gc.get_stats()
print(f"   集合总数现在: {stats_after['collection_count']}\n")

# 测试 5: 设置 GC 阈值
print("5. 测试 GC 阈值设置")
gc.set_threshold(2097152)  # 2MB
print("   GC 阈值已设置为 2MB\n")

# 测试 6: 设置详细模式
print("6. 测试详细模式")
gc.set_verbose(True)
print("   GC 详细模式已启用")
gc.set_verbose(False)
print("   GC 详细模式已禁用\n")

print("=== 垃圾回收测试完成 ===")
