import os

#for machine_index in range(8058, 8063):
os.system("ssh osgroup6@122.200.68.26 -p 8053 'nohup /osdata/osgroup6/final_project21_deploy/build/miner_five --log-difficulty=23 --my-ip=10.1.0.94 &' &")
os.system("ssh osgroup6@122.200.68.26 -p 8054 'nohup /osdata/osgroup6/final_project21_deploy/build/miner_five --log-difficulty=23 --my-ip=10.1.0.95 &' &")
os.system("ssh osgroup6@122.200.68.26 -p 8055 'nohup /osdata/osgroup6/final_project21_deploy/build/miner_five --log-difficulty=23 --my-ip=10.1.0.96 &' &")
os.system("ssh osgroup6@122.200.68.26 -p 8064 'nohup /osdata/osgroup6/final_project21_deploy/build/miner_five --log-difficulty=23 --my-ip=10.1.0.110 &' &")
os.system("ssh osgroup6@122.200.68.26 -p 8065 'nohup /osdata/osgroup6/final_project21_deploy/build/miner_five --log-difficulty=23 --my-ip=10.1.0.111 &' &")
os.system("ssh osgroup6@122.200.68.26 -p 8063 'nohup /osdata/osgroup6/final_project21_deploy/build/miner_five --log-difficulty=23 --my-ip=10.1.0.109 &' &")
os.system("ssh osgroup6@122.200.68.26 -p 8059 'nohup /osdata/osgroup6/final_project21_deploy/build/miner_five --log-difficulty=23 --my-ip=10.1.0.119 &' &")
os.system("ssh osgroup6@122.200.68.26 -p 8060 'nohup /osdata/osgroup6/final_project21_deploy/build/miner_five --log-difficulty=23 --my-ip=10.1.0.104 &' &")
os.system("ssh osgroup6@122.200.68.26 -p 8061 'nohup /osdata/osgroup6/final_project21_deploy/build/miner_five --log-difficulty=23 --my-ip=10.1.0.105 &' &")
os.system("ssh osgroup6@122.200.68.26 -p 8062 'nohup /osdata/osgroup6/final_project21_deploy/build/miner_five --log-difficulty=23 --my-ip=10.1.0.107 &' &")
for machine_index in range(8063, 8064):
    ...#os.system("ssh osgroup6@122.200.68.26 -p {} '/osdata/osgroup6/part1_test/build/client'".format(machine_index))