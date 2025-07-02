import socket
import subprocess
import re


def extract_operator(whois):
    provider_fields = [
        "OrgName",
        "Organization",
        "org-name",
        "descr",
        "owner",
        "netname",
        "responsible",
        "role",
        "person",
        "mnt-by",
    ]
    for field in provider_fields:
        match = re.search(rf"^{field}\s*:\s*(.+)$", whois, re.IGNORECASE | re.MULTILINE)
        if match:
            return match.group(1).strip()
    return "—"


def extract_country(whois):
    country_fields = ["country", "ctry", "co", "country-code"]
    for field in country_fields:
        match = re.search(
            rf"^{field}\s*:\s*(\w+)$", whois, re.IGNORECASE | re.MULTILINE
        )
        if match:
            return match.group(1).strip().upper()
    return "—"


def main():
    with open("in.txt", "r") as infd:
        domains = [line.strip() for line in infd if line.strip()]

    total = len(domains)

    with open("out.txt", "w") as outfd:
        outfd.write(f"|Domain|Provider|Country|\n")

        for idx, domain in enumerate(domains, start=1):
            provider = "—"
            country = "—"
            try:
                ip = socket.gethostbyname(domain)
                result = subprocess.run(
                    f"whois {ip}",
                    shell=True,
                    stdout=subprocess.PIPE,
                    stderr=subprocess.DEVNULL,
                    text=True,
                )
                provider = extract_operator(result.stdout)
                country = extract_country(result.stdout)
            except Exception:
                pass

            print(f"{idx}/{total}: {domain} => {provider} [{country}]")
            outfd.write(f"|{domain}|{provider}|{country}|\n")


if __name__ == "__main__":
    main()
