import os

for machine_index in range(8053, 8066):
    if machine_index == 8056 or machine_index == 8057 or machine_index == 8058:
        continue
    os.system("ssh osgroup6@122.200.68.26 -p {} 'mkdir -p /osdata/osgroup6/final_project21_deploy/build'".format(machine_index))
    os.system("scp -P {} ./build/miner_malicious1 osgroup6@122.200.68.26:/osdata/osgroup6/final_project21_deploy/build/".format(machine_index))
    os.system("scp -P {} ./build/miner_malicious2 osgroup6@122.200.68.26:/osdata/osgroup6/final_project21_deploy/build/".format(machine_index))
    os.system("scp -P {} ./build/miner_malicious3 osgroup6@122.200.68.26:/osdata/osgroup6/final_project21_deploy/build/".format(machine_index))
    os.system("scp -P {} ./build/miner_lazy osgroup6@122.200.68.26:/osdata/osgroup6/final_project21_deploy/build/".format(machine_index))
    os.system("scp -P {} ./build/miner_honest osgroup6@122.200.68.26:/osdata/osgroup6/final_project21_deploy/build/".format(machine_index))
    os.system("scp -P {} ./build/miner_fork osgroup6@122.200.68.26:/osdata/osgroup6/final_project21_deploy/build/".format(machine_index))
    os.system("scp -P {} ./build/client osgroup6@122.200.68.26:/osdata/osgroup6/final_project21_deploy/build/".format(machine_index))
    os.system("scp -P {} ./build/client_fork osgroup6@122.200.68.26:/osdata/osgroup6/final_project21_deploy/build/".format(machine_index))
    os.system("scp -P {} ./build/client_five osgroup6@122.200.68.26:/osdata/osgroup6/final_project21_deploy/build/".format(machine_index))
    os.system("scp -P {} ./build/miner osgroup6@122.200.68.26:/osdata/osgroup6/final_project21_deploy/build/".format(machine_index))
    os.system("scp -P {} ./build/miner_five osgroup6@122.200.68.26:/osdata/osgroup6/final_project21_deploy/build/".format(machine_index))