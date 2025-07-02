import subprocess
import datetime
import argparse

def log(msg):
    ts = datetime.datetime.now().strftime("%H:%M:%S.%f")[:-3]
    print(f"[{ts}] {msg}")


def main():
    parser = argparse.ArgumentParser(prog="Domain whitelist checker")

    parser.add_argument("-i", dest="in_path", default="in.txt", help="Input file path")
    parser.add_argument("-o", dest="out_path", default="out.txt", help="Output file path")
    parser.add_argument("-e", dest="err_path", default="err.txt", help="Error file path")
    parser.add_argument("-u", dest="url_path", required=True, help="URL path")
    parser.add_argument("-d", dest="dst_ip", required=True, help="Destination IP")
    parser.add_argument("-t", dest="timeout_sec", type=int, default=5, help="Connection/read timeout in seconds")
    parser.add_argument("-r", dest="range_to", type=int, default=65535, help="Upper bound of the range of bytes to be downloaded")
    
    args = parser.parse_args()

    with open(args.in_path, "r") as infd:
        items = [line.strip() for line in infd if line.strip()]

    total = len(items)

    with open(args.out_path, "w") as outfd, open(args.err_path, "w") as errfd:
        for idx, domain in enumerate(items, start=1):
            try:
                result = subprocess.getoutput(
                    f'curl -k https://{domain}{args.url_path} \
                        --resolve {domain}:443:{args.dst_ip} \
                        -o /dev/null \
                        -s -w "%{{size_download}}\n" \
                        --max-time {args.timeout_sec} \
                        --range 0-{args.range_to}'
                )

                bytes = float(result)
                if bytes >= args.range_to:
                    outfd.write(domain + "\n")

                log(f"[{idx}/{total}] {domain} => {bytes}")

            except Exception as e:
                errfd.write(domain + "\n")
                log(f"[{idx}/{total}] {domain} => ERROR: {e}")


if __name__ == "__main__":
    main()
