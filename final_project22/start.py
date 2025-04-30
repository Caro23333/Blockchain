import os
import subprocess
import argparse

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
my_index = 8056

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
parser.add_argument(
    "--num-miners-required",
    type=int,
    default=2,
    help="number of miners you required"
)
args = parser.parse_args()

is_client_started = False
client_index = -1
miner_indices = []
for machine_index in range(8051, 8071):
    if is_client_started and len(miner_indices) >= args.num_miners_required:
        break
    if machine_index == my_index:
        continue
    is_available = (subprocess.run(
        "ssh osgroup6@122.200.68.26 -p {} '{}; {}'"
        .format(
            machine_index,
            f"cd /osdata/osgroup6/{deployed_repo_name}/",
            f"build/net_tester --port={args.port} --port-client={args.port_client}"
        ),
        shell=True, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True
        ).stdout.strip() == "available")
    if is_available:
        if is_client_started:
            os.system(
                "nohup ssh osgroup6@122.200.68.26 -p {} '{}; {}' &"
                .format(
                    machine_index,
                    f"cd /osdata/osgroup6/{deployed_repo_name}/",
                    f"build/miner --my-ip={ip_mapping[machine_index]} --port={args.port} --port-client={args.port_client}"
                ))
            miner_indices.append(machine_index)
        else:
            os.system(
                "nohup ssh osgroup6@122.200.68.26 -p {} '{}; {}' &"
                .format(
                    machine_index,
                    f"cd /osdata/osgroup6/{deployed_repo_name}/",
                    f"build/client --my-ip={ip_mapping[machine_index]} --port={args.port} --port-client={args.port_client}"
                ))
            client_index = machine_index
            is_client_started = True

if client_index == -1:
    print(f"Sorry, there are no machines available at both port {args.port} and port {args.port_client}.",
        "Perhaps you can try later.")
else:
    print(f"Successfully started miners {miner_indices} and client {client_index}.")
    if len(miner_indices) < args.num_miners_required:
        print("Sorry the started miners are not enough, but all available machines have been used.")
    else:
        print("Available machines are enough for your requirement.")
