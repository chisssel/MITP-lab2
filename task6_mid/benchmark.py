import subprocess
import time
import os
import sys
import re
from pathlib import Path

SCRIPT_DIR = Path(__file__).parent
FASTAPI_DIR = SCRIPT_DIR / "fastapi_server"
GIN_DIR = SCRIPT_DIR / "gin_server"
AB_PATH = r"C:\Apache24\bin\ab.exe"

FASTAPI_URL = "http://127.0.0.1:8000"
GIN_URL = "http://127.0.0.1:8001"

REQUESTS = 10000
CONCURRENCY = 100
DURATION = "30s"

class Colors:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'

def print_header(text):
    print(f"\n{Colors.HEADER}{Colors.BOLD}{'='*50}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{text}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{'='*50}{Colors.ENDC}\n")

def print_server(name, status):
    color = Colors.OKGREEN if status else Colors.FAIL
    print(f"{color}[{name}]: {Colors.OKGREEN if status else Colors.FAIL}{'STARTED' if status else 'FAILED'}{Colors.ENDC}")

def run_command(cmd, cwd=None, shell=False):
    try:
        result = subprocess.run(
            cmd, 
            cwd=cwd, 
            shell=shell, 
            capture_output=True, 
            text=True,
            timeout=60
        )
        return result.returncode == 0, result.stdout, result.stderr
    except subprocess.TimeoutExpired:
        return False, "", "Timeout"
    except Exception as e:
        return False, "", str(e)

def check_server(url):
    success, stdout, _ = run_command(
        ["curl", "-s", "-o", "NUL", "-w", "%{http_code}", f"{url}/ping"],
        shell=True
    )
    return success and "200" in stdout

def wait_for_server(url, name, timeout=30):
    print(f"Waiting for {name}...", end=" ")
    for _ in range(timeout):
        if check_server(url):
            print(f"{Colors.OKGREEN}OK{Colors.ENDC}")
            return True
        time.sleep(0.5)
    print(f"{Colors.FAIL}FAILED{Colors.ENDC}")
    return False

def run_ab(url, endpoint, requests=REQUESTS, concurrency=CONCURRENCY):
    full_url = f"{url}{endpoint}"
    print(f"  Running: ab -n {requests} -c {concurrency} {full_url}")
    
    success, stdout, stderr = run_command(
        [AB_PATH, "-n", str(requests), "-c", str(concurrency), "-g", "ab_results.tsv", full_url],
        shell=True
    )
    
    if not success:
        print(f"  {Colors.FAIL}Error: {stderr[:200]}{Colors.ENDC}")
        return None
    
    return parse_ab_output(stdout)

def parse_ab_output(output):
    results = {}
    
    patterns = {
        "requests_per_sec": r"Requests per second:\s+([\d.]+)",
        "time_per_req": r"Time per request:\s+([\d.]+)",
        "mean": r"Mean:\s+([\d.]+)",
        "stddev": r"Stddev:\s+([\d.]+)",
        "50_percent": r"50%\s+([\d]+)",
        "66_percent": r"66%\s+([\d]+)",
        "75_percent": r"75%\s+([\d]+)",
        "90_percent": r"90%\s+([\d]+)",
        "95_percent": r"95%\s+([\d]+)",
        "98_percent": r"98%\s+([\d]+)",
        "99_percent": r"99%\s+([\d]+)",
        "100_percent": r"100%\s+([\d]+)",
        "failed": r"Failed requests:\s+([\d]+)",
        "complete": r"Complete requests:\s+([\d]+)",
    }
    
    for key, pattern in patterns.items():
        match = re.search(pattern, output)
        if match:
            results[key] = float(match.group(1))
    
    return results

def fmt(value, fmt_str):
    if value is None or value == 'N/A':
        return 'N/A'
    try:
        return f"{float(value):{fmt_str}}"
    except (ValueError, TypeError):
        return 'N/A'

def print_results(name, results):
    if not results:
        print(f"  {Colors.FAIL}No results{Colors.ENDC}")
        return
    
    print(f"\n  {Colors.OKCYAN}{name} Results:{Colors.ENDC}")
    print(f"    Requests/sec:    {fmt(results.get('requests_per_sec'), '.2f')}")
    print(f"    Time/req (mean): {fmt(results.get('mean'), '.3f')} ms")
    print(f"    50%:             {fmt(results.get('50_percent'), '.0f')} ms")
    print(f"    90%:             {fmt(results.get('90_percent'), '.0f')} ms")
    print(f"    99%:             {fmt(results.get('99_percent'), '.0f')} ms")
    print(f"    Failed:          {fmt(results.get('failed'), '.0f')}")

def main():
    print_header("FastAPI vs Gin Benchmark")
    
    print(f"{Colors.WARNING}Note: Make sure both servers are running before starting tests!{Colors.ENDC}")
    print(f"  - FastAPI should run on port 8000")
    print(f"  - Gin should run on port 8001")
    print()
    
    input(f"Press Enter when servers are ready...")
    
    print_header("Checking Servers")
    fastapi_ok = check_server(FASTAPI_URL)
    gin_ok = check_server(GIN_URL)
    print_server("FastAPI (port 8000)", fastapi_ok)
    print_server("Gin (port 8001)", gin_ok)
    
    if not fastapi_ok or not gin_ok:
        print(f"\n{Colors.FAIL}Error: One or more servers not responding{Colors.ENDC}")
        sys.exit(1)
    
    endpoints = [
        ("/ping", "Simple JSON response"),
        ("/json", "Complex JSON with UUID"),
        ("/slow", "50ms delayed response"),
    ]
    
    for endpoint, description in endpoints:
        print_header(f"GET {endpoint} - {description}")
        
        print(f"{Colors.OKBLUE}FastAPI:{Colors.ENDC}")
        fastapi_results = run_ab(FASTAPI_URL, endpoint)
        print_results("FastAPI", fastapi_results)
        
        print(f"\n{Colors.OKBLUE}Gin:{Colors.ENDC}")
        gin_results = run_ab(GIN_URL, endpoint)
        print_results("Gin", gin_results)
        
        if fastapi_results and gin_results:
            rps_f = fastapi_results.get("requests_per_sec", 0)
            rps_g = gin_results.get("requests_per_sec", 0)
            winner = "Gin" if rps_g > rps_f else "FastAPI"
            diff = abs(rps_g - rps_f) / max(rps_f, rps_g) * 100
            print(f"\n  {Colors.OKGREEN}Winner: {winner} (+{diff:.1f}% faster){Colors.ENDC}")
    
    print_header("/echo POST Test")
    print(f"{Colors.OKBLUE}FastAPI:{Colors.ENDC}")
    print(f"  Running: ab -n {REQUESTS} -c {CONCURRENCY} -p echo_data.txt -T application/json {FASTAPI_URL}/echo")
    
    echo_data_file = SCRIPT_DIR / "echo_data.txt"
    echo_data_file.write_text('{"key":"value","test":123}')
    
    success, stdout, stderr = run_command(
        [AB_PATH, "-n", str(REQUESTS), "-c", str(CONCURRENCY), 
         "-p", str(echo_data_file), "-T", "application/json", f"{FASTAPI_URL}/echo"],
        shell=True
    )
    if success:
        fastapi_echo = parse_ab_output(stdout)
        print_results("FastAPI", fastapi_echo)
    else:
        print(f"  {Colors.FAIL}Error{Colors.ENDC}")
    
    print(f"\n{Colors.OKBLUE}Gin:{Colors.ENDC}")
    success, stdout, stderr = run_command(
        [AB_PATH, "-n", str(REQUESTS), "-c", str(CONCURRENCY),
         "-p", str(echo_data_file), "-T", "application/json", f"{GIN_URL}/echo"],
        shell=True
    )
    if success:
        gin_echo = parse_ab_output(stdout)
        print_results("Gin", gin_echo)
    else:
        print(f"  {Colors.FAIL}Error{Colors.ENDC}")
    
    echo_data_file.unlink(missing_ok=True)
    
    print_header("Benchmark Complete!")
    print("Run tests again or analyze results above.")
    print(f"\nNote: Detailed results saved to ab_results.tsv if -g flag worked")

if __name__ == "__main__":
    main()
