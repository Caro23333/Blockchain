import os

os.system("ssh osgroup6@122.200.68.26 -p 8054 'nohup /osdata/osgroup6/final_project21_deploy/build/miner --my-ip=10.1.0.95 &' &")
os.system("ssh osgroup6@122.200.68.26 -p 8055 'nohup /osdata/osgroup6/final_project21_deploy/build/miner_malicious2 --my-ip=10.1.0.96 &' &")