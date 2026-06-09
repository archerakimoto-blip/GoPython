#!/usr/bin/env python3
"""
GoPython vs CPython 标准性能测试套件

全面测试 GoPython 解释器与 CPython 在各维度的性能表现对比。

测试类别:
  1. 算术运算 (arithmetic)
  2. 函数调用 (function)
  3. 循环 (loop)
  4. 数据结构 (datastructures)
  5. 面向对象 (oop)
  6. 字符串操作 (string)
  7. 异常处理 (exception)
  8. 生成器 (generator)
  9. 装饰器 (decorator)
  10. 内置函数 (builtins)
  11. 数学运算 (math)
  12. 变量与作用域 (scope)
  13. 综合应用 (app)

用法:
  python3 run_benchmarks.py                    # 运行全部基准测试
  python3 run_benchmarks.py --gopy /path/gopy  # 指定 gopy 路径
  python3 run_benchmarks.py --only arithmetic  # 只运行算术测试
  python3 run_benchmarks.py --repeat 3         # 每个测试重复3次取中位数
  python3 run_benchmarks.py --json result.json # 导出 JSON 结果
  python3 run_benchmarks.py --skip-cpython     # 跳过 CPython 基线
"""

import subprocess
import sys
import os
import re
import time
import json
import argparse
import statistics
from pathlib import Path
from collections import OrderedDict

# 基准测试文件定义: (文件名, 类别描述)
BENCHMARK_SUITES = OrderedDict([
    ("bench_arithmetic.py",     "算术运算"),
    ("bench_function.py",       "函数调用"),
    ("bench_loop.py",           "循环"),
    ("bench_datastructures.py", "数据结构"),
    ("bench_oop.py",            "面向对象"),
    ("bench_string.py",         "字符串操作"),
    ("bench_exception.py",      "异常处理"),
    ("bench_generator.py",      "生成器"),
    ("bench_decorator.py",      "装饰器"),
    ("bench_builtins.py",       "内置函数"),
    ("bench_math.py",           "数学运算"),
    ("bench_scope.py",          "变量与作用域"),
    ("bench_app.py",            "综合应用"),
])

# 结果解析正则
RESULT_PATTERN = re.compile(r"BENCHMARK_RESULT\|([^|]+)\|([0-9.]+)")

# 颜色输出
class Color:
    RESET  = "\033[0m"
    RED    = "\033[91m"
    GREEN  = "\033[92m"
    YELLOW = "\033[93m"
    BLUE   = "\033[94m"
    MAGENTA = "\033[95m"
    CYAN   = "\033[96m"
    BOLD   = "\033[1m"
    DIM    = "\033[2m"

    @staticmethod
    def supports_color():
        if os.environ.get("NO_COLOR"):
            return False
        if not hasattr(sys.stdout, "isatty"):
            return False
        return sys.stdout.isatty()

def c(color, text):
    if Color.supports_color():
        return f"{color}{text}{Color.RESET}"
    return text


def find_gopy():
    """查找 gopy 可执行文件"""
    # 1. 命令行指定
    # 2. 项目根目录
    candidates = [
        Path(__file__).parent.parent / "gopy",
        Path(__file__).parent / "gopy",
        Path.cwd() / "gopy",
    ]
    for p in candidates:
        if p.exists() and os.access(p, os.X_OK):
            return str(p)
    # 3. PATH
    import shutil
    gopy = shutil.which("gopy")
    if gopy:
        return gopy
    return None


def run_single_benchmark(runner_cmd, script_path, timeout=120):
    """运行单个基准测试脚本，返回 {name: time} 字典"""
    results = {}
    try:
        proc = subprocess.run(
            runner_cmd + [script_path],
            capture_output=True,
            text=True,
            timeout=timeout,
        )
        for line in proc.stdout.splitlines():
            m = RESULT_PATTERN.match(line.strip())
            if m:
                name = m.group(1)
                val = float(m.group(2))
                results[name] = val
        # 检查是否有错误
        if proc.returncode != 0 and not results:
            stderr = proc.stderr.strip()
            if stderr:
                return {"__error__": stderr[:200]}
    except subprocess.TimeoutExpired:
        return {"__error__": "TIMEOUT"}
    except Exception as e:
        return {"__error__": str(e)[:200]}
    return results


def run_benchmark_with_repeats(runner_cmd, script_path, repeat=1, timeout=120):
    """多次运行取中位数"""
    all_runs = []
    for _ in range(repeat):
        result = run_single_benchmark(runner_cmd, script_path, timeout)
        if "__error__" in result:
            return result
        all_runs.append(result)

    # 合并: 每个指标取中位数
    merged = {}
    keys = all_runs[0].keys()
    for key in keys:
        values = [run[key] for run in all_runs]
        merged[key] = statistics.median(values)
    return merged


def format_time(seconds):
    """格式化时间显示"""
    if seconds < 0.001:
        return f"{seconds * 1000000:.1f} us"
    elif seconds < 1:
        return f"{seconds * 1000:.2f} ms"
    else:
        return f"{seconds:.3f} s"


def format_ratio(ratio):
    """格式化比率，带颜色"""
    if ratio is None:
        return c(Color.DIM, "N/A")
    if ratio >= 1:
        # GoPython 更慢
        if ratio >= 10:
            return c(Color.RED, f"{ratio:.1f}x slower")
        elif ratio >= 3:
            return c(Color.YELLOW, f"{ratio:.1f}x slower")
        elif ratio >= 1.5:
            return c(Color.YELLOW, f"{ratio:.1f}x slower")
        else:
            return f"{ratio:.2f}x slower"
    else:
        # GoPython 更快
        speedup = 1 / ratio
        if speedup >= 10:
            return c(Color.GREEN, f"{speedup:.1f}x faster")
        elif speedup >= 3:
            return c(Color.GREEN, f"{speedup:.1f}x faster")
        else:
            return c(Color.GREEN, f"{speedup:.2f}x faster")


def print_header():
    """打印标题"""
    print()
    print(c(Color.BOLD + Color.CYAN, "=" * 80))
    print(c(Color.BOLD + Color.CYAN, "  GoPython vs CPython 标准性能测试套件"))
    print(c(Color.BOLD + Color.CYAN, "=" * 80))
    print()


def print_system_info(gopy_path, cpython_path):
    """打印系统信息"""
    import platform
    print(c(Color.BOLD, "系统信息:"))
    print(f"  OS:        {platform.system()} {platform.release()}")
    print(f"  Arch:      {platform.machine()}")
    print(f"  CPython:   {platform.python_version()}")

    # GoPython 版本
    if gopy_path:
        print(f"  GoPython:  {gopy_path}")
    else:
        print(f"  GoPython:  " + c(Color.RED, "NOT FOUND"))

    print()


def print_category_header(name, description):
    """打印类别标题"""
    print(c(Color.BOLD + Color.BLUE, f"\n[{name}] {description}"))
    print(c(Color.DIM, "-" * 70))


def print_result_row(name, gopy_time, cpython_time, status="ok"):
    """打印单行结果"""
    name_col = f"  {name:<35}"
    if status == "error":
        print(f"{name_col} {c(Color.RED, 'ERROR')}")
        return
    if status == "skip":
        print(f"{name_col} {c(Color.DIM, 'SKIPPED')}")
        return

    gopy_col = f"{format_time(gopy_time):>12}"
    cpython_col = f"{format_time(cpython_time):>12}" if cpython_time is not None else f"{'N/A':>12}"

    ratio = None
    if cpython_time and cpython_time > 0:
        ratio = gopy_time / cpython_time

    ratio_col = format_ratio(ratio)

    print(f"{name_col} {gopy_col}  {cpython_col}  {ratio_col}")


def print_summary(all_results):
    """打印汇总统计"""
    print()
    print(c(Color.BOLD + Color.CYAN, "=" * 80))
    print(c(Color.BOLD + Color.CYAN, "  汇总统计"))
    print(c(Color.BOLD + Color.CYAN, "=" * 80))

    total_gopy = 0
    total_cpython = 0
    count = 0
    faster_count = 0
    slower_count = 0
    ratios = []

    for bench_name, results in all_results.items():
        for metric, data in results.items():
            if data.get("gopy") is not None and data.get("cpython") is not None:
                gopy_t = data["gopy"]
                cpython_t = data["cpython"]
                if cpython_t > 0:
                    ratio = gopy_t / cpython_t
                    ratios.append(ratio)
                    if ratio < 1:
                        faster_count += 1
                    else:
                        slower_count += 1
                    total_gopy += gopy_t
                    total_cpython += cpython_t
                    count += 1

    if count == 0:
        print(c(Color.YELLOW, "  无可比较的数据"))
        return

    print()
    print(f"  总测试项:          {count}")
    print(f"  GoPython 更快:     {c(Color.GREEN, str(faster_count))} 项")
    print(f"  GoPython 更慢:     {c(Color.RED, str(slower_count))} 项")

    if ratios:
        geo_mean = statistics.geometric_mean(ratios)
        median_ratio = statistics.median(ratios)
        print(f"  几何平均比率:      {format_ratio(geo_mean)}")
        print(f"  中位数比率:        {format_ratio(median_ratio)}")

        # 按类别统计
        print()
        print(c(Color.BOLD, "  按类别统计:"))
        category_stats = OrderedDict()
        for bench_name, results in all_results.items():
            cat_ratios = []
            for metric, data in results.items():
                if data.get("gopy") is not None and data.get("cpython") is not None:
                    cpython_t = data["cpython"]
                    if cpython_t > 0:
                        cat_ratios.append(data["gopy"] / cpython_t)
            if cat_ratios:
                category_stats[bench_name] = statistics.geometric_mean(cat_ratios)

        for cat, ratio in category_stats.items():
            desc = BENCHMARK_SUITES.get(cat, cat)
            print(f"    {desc:<16} {format_ratio(ratio)}")

    print()

    # 最慢和最快的 Top 5
    if ratios:
        sorted_items = []
        for bench_name, results in all_results.items():
            for metric, data in results.items():
                if data.get("gopy") is not None and data.get("cpython") is not None:
                    cpython_t = data["cpython"]
                    if cpython_t > 0:
                        sorted_items.append((metric, data["gopy"] / cpython_t))

        sorted_items.sort(key=lambda x: x[1])

        print(c(Color.BOLD, "  GoPython 最快的 5 项:"))
        for name, ratio in sorted_items[:5]:
            print(f"    {name:<35} {format_ratio(ratio)}")

        print()
        print(c(Color.BOLD, "  GoPython 最慢的 5 项:"))
        for name, ratio in sorted_items[-5:]:
            print(f"    {name:<35} {format_ratio(ratio)}")

    print()


def export_json(all_results, output_path):
    """导出结果为 JSON"""
    export_data = {
        "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
        "benchmarks": {}
    }
    for bench_name, results in all_results.items():
        export_data["benchmarks"][bench_name] = {}
        for metric, data in results.items():
            entry = {}
            if data.get("gopy") is not None:
                entry["gopy_time"] = data["gopy"]
            if data.get("cpython") is not None:
                entry["cpython_time"] = data["cpython"]
            if "gopy_time" in entry and "cpython_time" in entry and entry["cpython_time"] > 0:
                entry["ratio"] = entry["gopy_time"] / entry["cpython_time"]
            if data.get("error"):
                entry["error"] = data["error"]
            export_data["benchmarks"][bench_name][metric] = entry

    with open(output_path, "w", encoding="utf-8") as f:
        json.dump(export_data, f, indent=2, ensure_ascii=False)
    print(f"结果已导出到: {output_path}")


def main():
    parser = argparse.ArgumentParser(
        description="GoPython vs CPython 标准性能测试套件"
    )
    parser.add_argument(
        "--gopy", default=None,
        help="GoPython 可执行文件路径"
    )
    parser.add_argument(
        "--repeat", type=int, default=1,
        help="每个测试重复次数 (取中位数, 默认1)"
    )
    parser.add_argument(
        "--only", default=None,
        help="只运行指定类别的测试 (如 arithmetic, function 等)"
    )
    parser.add_argument(
        "--skip-cpython", action="store_true",
        help="跳过 CPython 基线测试"
    )
    parser.add_argument(
        "--skip-gopy", action="store_true",
        help="跳过 GoPython 测试"
    )
    parser.add_argument(
        "--json", default=None,
        help="导出结果为 JSON 文件"
    )
    parser.add_argument(
        "--timeout", type=int, default=120,
        help="单个测试超时时间 (秒, 默认120)"
    )
    args = parser.parse_args()

    # 确定 gopy 路径
    gopy_path = args.gopy or find_gopy()

    print_header()
    print_system_info(gopy_path, sys.executable)

    # 表头
    print(c(Color.BOLD, f"  {'测试项':<35} {'GoPython':>12}  {'CPython':>12}  {'对比'}"))
    print(c(Color.DIM, "-" * 70))

    bench_dir = Path(__file__).parent
    all_results = OrderedDict()

    # 确定要运行的测试
    suites_to_run = OrderedDict()
    for filename, desc in BENCHMARK_SUITES.items():
        if args.only:
            # 匹配文件名前缀或类别描述
            if args.only.lower() in filename.lower() or args.only.lower() in desc.lower():
                suites_to_run[filename] = desc
        else:
            suites_to_run[filename] = desc

    if not suites_to_run:
        print(c(Color.RED, "没有匹配的测试类别"))
        sys.exit(1)

    total_benchmarks = len(suites_to_run)
    completed = 0

    for filename, desc in suites_to_run.items():
        script_path = str(bench_dir / filename)
        if not os.path.exists(script_path):
            print_category_header(filename, desc)
            print(c(Color.YELLOW, f"  文件不存在: {script_path}"))
            continue

        print_category_header(filename, desc)

        results = OrderedDict()

        # 运行 GoPython
        gopy_results = {}
        if not args.skip_gopy and gopy_path:
            gopy_results = run_benchmark_with_repeats(
                [gopy_path], script_path, args.repeat, args.timeout
            )

        # 运行 CPython
        cpython_results = {}
        if not args.skip_cpython:
            cpython_results = run_benchmark_with_repeats(
                [sys.executable], script_path, args.repeat, args.timeout
            )

        # 合并结果
        all_keys = list(dict.fromkeys(
            list(gopy_results.keys()) + list(cpython_results.keys())
        ))

        for key in all_keys:
            if key == "__error__":
                continue

            gopy_time = gopy_results.get(key)
            cpython_time = cpython_results.get(key)

            gopy_error = None
            cpython_error = None

            if gopy_results.get("__error__"):
                gopy_error = gopy_results["__error__"]
            if cpython_results.get("__error__"):
                cpython_error = cpython_results["__error__"]

            if gopy_error and not gopy_time:
                print_result_row(key, None, None, "error")
                print(c(Color.RED, f"    GoPython 错误: {gopy_error}"))
                results[key] = {"gopy": None, "cpython": cpython_time, "error": gopy_error}
            elif cpython_error and not cpython_time:
                print_result_row(key, gopy_time, None)
                results[key] = {"gopy": gopy_time, "cpython": None, "error": cpython_error}
            else:
                print_result_row(key, gopy_time, cpython_time)
                results[key] = {"gopy": gopy_time, "cpython": cpython_time}

        all_results[filename] = results
        completed += 1

    # 汇总
    print_summary(all_results)

    # 导出 JSON
    if args.json:
        export_json(all_results, args.json)


if __name__ == "__main__":
    main()
