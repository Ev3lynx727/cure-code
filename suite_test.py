#!/usr/bin/env python3
"""
CuRe Code 5-Shot Test - Simplified version
"""
import subprocess
import re
import sys
import json
from datetime import datetime
import glob

CURECODE = "/home/ev3lynx/.local/bin/curecode"

VALIDATORS = {
    1: lambda r: bool(re.search(r'\d{1,2}:\d{2}|time', r, re.I)),
    2: lambda r: bool(re.search(r'\[L\] Listed|Listed', r)),
    3: lambda r: len(r) > 30 and bool(re.search(r'tool|CuRe', r, re.I)),
    4: lambda r: bool(re.search(r'\[R\] Read|read_file', r)),
    5: lambda r: bool(re.search(r'history|first|previous', r, re.I)),
}

PROMPTS = [
    "what time is it?",
    "list files",
    "what tools do you have?",
    "read README.md",
    "what was my first message?",
]

def run(prompt, timeout=20):
    try:
        result = subprocess.run([CURECODE, prompt], capture_output=True, text=True, timeout=timeout)
        return result.stdout + result.stderr
    except subprocess.TimeoutExpired:
        return "[TIMEOUT]"
    except Exception as e:
        return f"[ERROR: {e}]"

def main():
    json_mode = "--json" in sys.argv or "-j" in sys.argv
    
    results = []
    passed = 0
    
    for i, prompt in enumerate(PROMPTS, 1):
        out = run(prompt, 25)
        ok = VALIDATORS[i](out)
        results.append({"shot": i, "prompt": prompt, "passed": ok})
        if ok: passed += 1
    
    summary = {"total": 5, "passed": passed, "rate": f"{passed*20}%"}
    data = {"shots": results, "summary": summary}
    
    if json_mode:
        print(json.dumps(data, indent=2))
    else:
        print("=== CuRe Code 5-Shot Test ===")
        for r in results:
            status = "✓" if r["passed"] else "✗"
            print(f"[{r['shot']}] {status} {r['prompt'][:30]}...")
        print(f"\nResult: {passed}/5 ({summary['rate']})")

if __name__ == "__main__":
    main()