import os

os.system("ssh osgroup6@122.200.68.26 -p 8053 'nohup /osdata/osgroup6/final_project21_deploy/build/miner_honest --my-ip=10.1.0.94 &' &")
os.system("ssh osgroup6@122.200.68.26 -p 8054 'nohup /osdata/osgroup6/final_project21_deploy/build/miner_fork --my-ip=10.1.0.95 &' &")
os.system("ssh osgroup6@122.200.68.26 -p 8055 'nohup /osdata/osgroup6/final_project21_deploy/build/miner_honest --my-ip=10.1.0.96 &' &")