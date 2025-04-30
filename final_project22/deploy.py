import os

# fold structure: final_project22_deploy { build { server , client , helper } , temp { command.txt , result.txt } }
# file form: command.txt = "12 query ####", result.txt = "12 17,60 18,70 ####"

deployed_repo_name = "final_project22_deploy"
my_index = 8056
empty_symbol = "#-empty-#"

for machine_index in range(8051, 8071):
    if machine_index == my_index:
        continue
    os.system(
        "ssh osgroup6@122.200.68.26 -p {} '{}; {}; {}; {}; {}; {}'"
        .format(
            machine_index,
            f"rm -rf /osdata/osgroup6/{deployed_repo_name}/",
            f"mkdir -p /osdata/osgroup6/{deployed_repo_name}/",
            f"cd /osdata/osgroup6/{deployed_repo_name}/",
            "mkdir -p temp/",
            f"echo -n \"{empty_symbol}\" > temp/command.txt",
            f"echo -n \"{empty_symbol}\" > temp/result.txt"
        ))
    os.system(
        "scp -r -P {} build/ osgroup6@122.200.68.26:/osdata/osgroup6/{}/"
        .format(machine_index, deployed_repo_name))
