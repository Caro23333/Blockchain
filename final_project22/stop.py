import argparse
import os
import subprocess

ip_mapping = {
	8051: "10.1.0.91",
	8052: "10.1.0.92",
	8053: "10.1.0.94",
	8054: "10.1.0.95",
	8055: "10.1.0.96",
    8056: "10.1.0.98",
	8057: "10.1.0.99",
    8058: "10.1.0.102",
    8059: "10.1.0.119",
    8060: "10.1.0.104",
    8061: "10.1.0.105",
	8062: "10.1.0.107",
	8063: "10.1.0.109",
	8064: "10.1.0.110",
	8065: "10.1.0.111",
	8066: "10.1.0.112",
	8067: "10.1.0.113",
	8068: "10.1.0.115",
	8069: "10.1.0.116",
    8070: "10.1.0.117",
}
deployed_repo_name = "final_project22_deploy"

parser = argparse.ArgumentParser()
parser.add_argument(
    "--port",
    type=str,
    default="8056",
    help="port of connection between miners"
)
parser.add_argument(
    "--port-client",
    type=str,
    default="8057",
    help="port of connection between miners and client"
)
args = parser.parse_args()

# here we just roughly filter processes
# below cannot add "--" before "port" and "port-client", otherwise `fgrep` will raise error
processes = subprocess.run(
    "ps aux | fgrep '{}' | fgrep '{}' | fgrep '{}' | fgrep '{}' | fgrep '{}'".format(
        "osgroup6",
        "ssh osgroup6@122.200.68.26",
        f"cd /osdata/osgroup6/{deployed_repo_name}/",
        f"port={args.port}",
        f"port-client={args.port_client}"
    ),
    shell=True, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True
).stdout.strip().split('\n')

# here we precisely filter processes
had_running_service = False
for process in processes:
    words = [item for item in process.split(' ') if item != ""]
    pid = words[1]
    if "ssh" not in words:
        continue
    command = words[words.index("ssh"):]
    if len(command) < 10:
        continue
    machine_index = command[3]
    if '/' not in command[6]:
        continue
    service_type = command[6].split('/')[1]
    if service_type != "miner" and service_type != "client":
        continue
    expected_command_str = "ssh osgroup6@122.200.68.26 -p {} {}; {}".format(
        machine_index,
        f"cd /osdata/osgroup6/{deployed_repo_name}/",
        f"build/{service_type} --my-ip={ip_mapping[int(machine_index)]} --port={args.port} --port-client={args.port_client}"
    )
    if ' '.join(command) != expected_command_str:
        continue
    had_running_service = True
    os.system(f"kill {pid}")
    print(f"Successfully killed {service_type} at {machine_index}.")
if not had_running_service:
    print(f"There are no miners or clients running at port = {args.port} and port-client = {args.port_client}.")
