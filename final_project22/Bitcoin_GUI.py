import sys
import subprocess
from PyQt5.QtWidgets import QApplication, QWidget, QVBoxLayout, QHBoxLayout, QGridLayout
from PyQt5.QtWidgets import QPushButton, QLineEdit, QLabel
from PyQt5.QtWidgets import QMessageBox, QDialog, QTableWidget, QTableWidgetItem
from PyQt5.QtGui import QIntValidator, QDoubleValidator, QFont
from PyQt5.QtCore import Qt

public_ip = "122.200.68.26"
client_port = "8070"
username = "osgroup6"
deployed_repo_name = "updated_gui_project22_deploy"

def isNonnegativeInteger(s: str) -> bool:
    try:
        x = int(s)
        return (x >= 0)
    except ValueError:
        return False

def isNonnegativeFloat(s: str) -> bool:
    try:
        x = float(s)
        return (x >= 0)
    except ValueError:
        return False

def approx(s: str, n: int = 2) -> str:
    a = str(int(float(s) * 10 ** n + 0.5))
    if len(a) <= n:
        a = "0" * (n + 1 - len(a)) + a
    return a[ : len(a) - n] + "." + a[len(a) - n : ]

class SettingsDialog(QDialog):
    def __init__(self):
        super().__init__()

        self.setWindowTitle("Settings")
        self.setGeometry(200, 200, 800, 800)

        # Create labels and input fields
        self.public_ip_label = QLabel("Public IP:")
        self.public_ip_input = QLineEdit(public_ip)

        self.client_index_label = QLabel("Client Port:")
        self.client_index_input = QLineEdit(client_port)

        self.username_label = QLabel("Username:")
        self.username_input = QLineEdit(username)

        # Create button
        self.do_set_button = QPushButton("Set")
        self.do_set_button.clicked.connect(self.doSet)
 
        # Layout setup
        self.argument_layout = QGridLayout()
        self.argument_layout.addWidget(self.public_ip_label, 0, 0, 1, 1, alignment=Qt.AlignRight)
        self.argument_layout.addWidget(self.public_ip_input, 0, 1, 1, 3)
        self.argument_layout.addWidget(self.client_index_label, 1, 0, 1, 1, alignment=Qt.AlignRight)
        self.argument_layout.addWidget(self.client_index_input, 1, 1, 1, 3)
        self.argument_layout.addWidget(self.username_label, 2, 0, 1, 1, alignment=Qt.AlignRight)
        self.argument_layout.addWidget(self.username_input, 2, 1, 1, 3)
        
        self.all_layout = QVBoxLayout()
        self.all_layout.addLayout(self.argument_layout)
        self.all_layout.addWidget(self.do_set_button)

        self.all_widget = QWidget()
        self.all_widget.setFixedSize(800, 400)
        self.all_widget.setLayout(self.all_layout)

        self.layout = QVBoxLayout()
        self.layout.addWidget(self.all_widget, alignment=Qt.AlignCenter)

        self.setLayout(self.layout)

    # No type-safe check here, because it is relatively irrelevant to the project, and needs much time (e.g. ssh test, public key GUI, etc.)
    # So please do not be malicious or make mistakes here
    def doSet(self):
        global public_ip, client_port, username
        public_ip = self.public_ip_input.text()
        client_port = self.client_index_input.text()
        username = self.username_input.text()
        QMessageBox.information(self, "Setting Result", "Successfully sets!")

class QueryResultDialog(QDialog):
    def __init__(self, query_result_list: list[str]):
        super().__init__()

        self.setWindowTitle("Querying Balance Result")
        self.setGeometry(200, 200, 800, 800)

        # 创建一个表格小部件
        self.table = QTableWidget()
        self.table.setRowCount(len(query_result_list))
        self.table.setColumnCount(2)
        self.table.setColumnWidth(0, 300)
        self.table.setColumnWidth(1, 300)
        self.table.setHorizontalHeaderLabels(["index", "balance"])
        self.table.verticalHeader().hide()
 
        # 填充表格数据
        for index, item in enumerate(query_result_list):
            self.table.setItem(index, 0, QTableWidgetItem(item.split(',')[0]))
            self.table.setItem(index, 1, QTableWidgetItem(approx(item.split(',')[1])))
 
        # 创建一个布局并将表格添加到其中
        layout = QVBoxLayout()
        layout.addWidget(self.table)
        self.setLayout(layout)

class TxDialog(QDialog):
    def __init__(self):
        super().__init__()

        self.setWindowTitle("Do Transaction")
        self.setGeometry(200, 200, 800, 800)

        # Create labels and input fields
        self.source_label = QLabel("Source:")
        self.source_input = QLineEdit()
        self.source_input.setValidator(QIntValidator())

        self.dest_label = QLabel("Dest:")
        self.dest_input = QLineEdit()
        self.dest_input.setValidator(QIntValidator())

        self.value_label = QLabel("Value:")
        self.value_input = QLineEdit()
        self.value_input.setValidator(QDoubleValidator(0.01, 999999, 2))  # Allow positive real numbers with up to 2 decimals

        self.fee_label = QLabel("Fee:")
        self.fee_input = QLineEdit()
        self.fee_input.setValidator(QDoubleValidator(0.01, 999999, 2))  # Allow positive real numbers with up to 2 decimals

        # Create button
        self.do_tx_button = QPushButton("Do Transaction")
        self.do_tx_button.clicked.connect(self.doTx)
 
        # Layout setup
        self.argument_layout = QGridLayout()
        self.argument_layout.addWidget(self.source_label, 0, 0, 1, 1, alignment=Qt.AlignRight)
        self.argument_layout.addWidget(self.source_input, 0, 1, 1, 3)
        self.argument_layout.addWidget(self.dest_label, 1, 0, 1, 1, alignment=Qt.AlignRight)
        self.argument_layout.addWidget(self.dest_input, 1, 1, 1, 3)
        self.argument_layout.addWidget(self.value_label, 2, 0, 1, 1, alignment=Qt.AlignRight)
        self.argument_layout.addWidget(self.value_input, 2, 1, 1, 3)
        self.argument_layout.addWidget(self.fee_label, 3, 0, 1, 1, alignment=Qt.AlignRight)
        self.argument_layout.addWidget(self.fee_input, 3, 1, 1, 3)
        
        self.all_layout = QVBoxLayout()
        self.all_layout.addLayout(self.argument_layout)
        self.all_layout.addWidget(self.do_tx_button)

        self.all_widget = QWidget()
        self.all_widget.setFixedSize(800, 500)
        self.all_widget.setLayout(self.all_layout)

        self.layout = QVBoxLayout()
        self.layout.addWidget(self.all_widget, alignment=Qt.AlignCenter)

        self.setLayout(self.layout)
    
    def doTx(self):
        source = self.source_input.text()
        dest = self.dest_input.text()
        value = self.value_input.text()
        fee = self.fee_input.text()

        if not isNonnegativeInteger(source):
            QMessageBox.warning(self, "Invalid Input", "source should be a nonnegative integer.")
            return

        if not isNonnegativeInteger(dest):
            QMessageBox.warning(self, "Invalid Input", "dest should be a nonnegative integer.")
            return

        if not isNonnegativeFloat(value):
            QMessageBox.warning(self, "Invalid Input", "value should be a nonnegative float.")
            return

        if not isNonnegativeFloat(fee):
            QMessageBox.warning(self, "Invalid Input", "fee should be a nonnegative float.")
            return

        # Simulate SSH command execution (replace with actual command)
        # Here we just use a simple example to show the output
        cmd = f"ssh -p {client_port} {username}@{public_ip} \" \
            cd /osdata/osgroup6/{deployed_repo_name}/; \
            build/helper --command='tx {source} {dest} {value} {fee}' \
        \""
        try:
            result = subprocess.run(cmd, shell=True, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
            QMessageBox.information(self, "Doing Transaction Result", result.stdout.strip())
        except subprocess.CalledProcessError as e:
            QMessageBox.critical(self, f"Doing Transaction Error", f"An error occurred: {e.stderr.strip()}")

class AdjustDialog(QDialog):
    def __init__(self):
        super().__init__()

        self.setWindowTitle("Adjust Difficulty")
        self.setGeometry(200, 200, 800, 800)

        # Create labels and input fields
        self.log_difficulty_label = QLabel("Log Difficulty:")
        self.log_difficulty_input = QLineEdit()
        self.log_difficulty_input.setValidator(QIntValidator())

        # Create button
        self.do_adjust_button = QPushButton("Adjust Difficulty")
        self.do_adjust_button.clicked.connect(self.doAdjust)
 
        # Layout setup
        self.argument_layout = QGridLayout()
        self.argument_layout.addWidget(self.log_difficulty_label, 0, 0, 1, 1, alignment=Qt.AlignRight)
        self.argument_layout.addWidget(self.log_difficulty_input, 0, 1, 1, 3)
        
        self.all_layout = QVBoxLayout()
        self.all_layout.addLayout(self.argument_layout)
        self.all_layout.addWidget(self.do_adjust_button)

        self.all_widget = QWidget()
        self.all_widget.setFixedSize(800, 200)
        self.all_widget.setLayout(self.all_layout)

        self.layout = QVBoxLayout()
        self.layout.addWidget(self.all_widget, alignment=Qt.AlignCenter)

        self.setLayout(self.layout)
    
    def doAdjust(self):
        log_difficulty = self.log_difficulty_input.text()

        if not isNonnegativeInteger(log_difficulty):
            QMessageBox.warning(self, "Invalid Input", "log_difficulty should be a nonnegative integer.")
            return

        # Simulate SSH command execution (replace with actual command)
        # Here we just use a simple example to show the output
        cmd = f"ssh -p {client_port} {username}@{public_ip} \" \
            cd /osdata/osgroup6/{deployed_repo_name}/; \
            build/helper --command='adjust {log_difficulty}' \
        \""
        try:
            result = subprocess.run(cmd, shell=True, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
            QMessageBox.information(self, "Adjusting Difficulty Result", result.stdout.strip())
        except subprocess.CalledProcessError as e:
            QMessageBox.critical(self, f"Adjusting Difficulty Error", f"An error occurred: {e.stderr.strip()}")

class ConnectDialog(QDialog):
    def __init__(self):
        super().__init__()

        self.setWindowTitle("Connect Miners")
        self.setGeometry(200, 200, 800, 800)

        # Create labels and input fields
        self.source_label = QLabel("Source:")
        self.source_input = QLineEdit()
        self.source_input.setValidator(QIntValidator())

        self.dest_label = QLabel("Dest:")
        self.dest_input = QLineEdit()
        self.dest_input.setValidator(QIntValidator())

        # Create button
        self.do_connect_button = QPushButton("Connect Miners")
        self.do_connect_button.clicked.connect(self.doConnect)
 
        # Layout setup
        self.argument_layout = QGridLayout()
        self.argument_layout.addWidget(self.source_label, 0, 0, 1, 1, alignment=Qt.AlignRight)
        self.argument_layout.addWidget(self.source_input, 0, 1, 1, 3)
        self.argument_layout.addWidget(self.dest_label, 1, 0, 1, 1, alignment=Qt.AlignRight)
        self.argument_layout.addWidget(self.dest_input, 1, 1, 1, 3)
        
        self.all_layout = QVBoxLayout()
        self.all_layout.addLayout(self.argument_layout)
        self.all_layout.addWidget(self.do_connect_button)

        self.all_widget = QWidget()
        self.all_widget.setFixedSize(800, 300)
        self.all_widget.setLayout(self.all_layout)

        self.layout = QVBoxLayout()
        self.layout.addWidget(self.all_widget, alignment=Qt.AlignCenter)

        self.setLayout(self.layout)
    
    def doConnect(self):
        source = self.source_input.text()
        dest = self.dest_input.text()

        if not isNonnegativeInteger(source):
            QMessageBox.warning(self, "Invalid Input", "source should be a nonnegative integer.")
            return

        if not isNonnegativeInteger(dest):
            QMessageBox.warning(self, "Invalid Input", "dest should be a nonnegative integer.")
            return

        # Simulate SSH command execution (replace with actual command)
        # Here we just use a simple example to show the output
        cmd = f"ssh -p {client_port} {username}@{public_ip} \" \
            cd /osdata/osgroup6/{deployed_repo_name}/; \
            build/helper --command='connect {source} {dest}' \
        \""
        try:
            result = subprocess.run(cmd, shell=True, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
            QMessageBox.information(self, "Connecting Miners Result", result.stdout.strip())
        except subprocess.CalledProcessError as e:
            QMessageBox.critical(self, f"Connecting Miners Error", f"An error occurred: {e.stderr.strip()}")

class NewnodeDialog(QDialog):
    def __init__(self):
        super().__init__()

        self.setWindowTitle("Add New Miner")
        self.setGeometry(200, 200, 800, 800)

        # Create labels and input fields
        self.index_label = QLabel("Index:")
        self.index_input = QLineEdit()
        self.index_input.setValidator(QIntValidator())

        # Create button
        self.do_newnode_button = QPushButton("Add New Miner")
        self.do_newnode_button.clicked.connect(self.doNewnode)
 
        # Layout setup
        self.argument_layout = QGridLayout()
        self.argument_layout.addWidget(self.index_label, 0, 0, 1, 1, alignment=Qt.AlignRight)
        self.argument_layout.addWidget(self.index_input, 0, 1, 1, 3)
        
        self.all_layout = QVBoxLayout()
        self.all_layout.addLayout(self.argument_layout)
        self.all_layout.addWidget(self.do_newnode_button)

        self.all_widget = QWidget()
        self.all_widget.setFixedSize(800, 200)
        self.all_widget.setLayout(self.all_layout)

        self.layout = QVBoxLayout()
        self.layout.addWidget(self.all_widget, alignment=Qt.AlignCenter)

        self.setLayout(self.layout)
    
    def doNewnode(self):
        index = self.index_input.text()

        if not isNonnegativeInteger(index):
            QMessageBox.warning(self, "Invalid Input", "index should be a nonnegative integer.")
            return

        # Simulate SSH command execution (replace with actual command)
        # Here we just use a simple example to show the output
        cmd = f"ssh -p {client_port} {username}@{public_ip} \" \
            cd /osdata/osgroup6/{deployed_repo_name}/; \
            build/helper --command='newnode {index}' \
        \""
        try:
            result = subprocess.run(cmd, shell=True, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
            QMessageBox.information(self, "Adding New Miner Result", result.stdout.strip())
        except subprocess.CalledProcessError as e:
            QMessageBox.critical(self, f"Adding New Miner Error", f"An error occurred: {e.stderr.strip()}")

class BitcoinGUI(QWidget):
    def __init__(self):
        super().__init__()

        self.setWindowTitle("Bitcoin GUI")
        self.setGeometry(100, 100, 1500, 1200)

        self.declareIdentificationButtons()

        self.setIdentificationButtonsLayoutWidget()

        self.layout = QVBoxLayout()
        self.layout.addWidget(self.identification_widget, alignment = Qt.AlignCenter)

        self.setLayout(self.layout)

    def clearAllWidgets(self):
        while self.layout.count():
            item = self.layout.takeAt(0)
            widget = item.widget()
            if widget is not None:
                widget.deleteLater()
    
    def declareIdentificationButtons(self):
        self.user_button = QPushButton("I am User", self)
        self.user_button.clicked.connect(self.toUser)
        
        self.admin_button = QPushButton("I am Admin", self)
        self.admin_button.clicked.connect(self.toAdmin)

    def declareServiceButtons(self, is_admin: bool):
        self.do_query_button = QPushButton("Query Miners' Balance", self)
        self.do_query_button.clicked.connect(self.doQuery)

        self.tx_button = QPushButton("Do Transaction", self)
        self.tx_button.clicked.connect(self.toTxDialog)

        if is_admin:
            self.adjust_button = QPushButton("Adjust Difficulty", self)
            self.adjust_button.clicked.connect(self.toAdjustDialog)

            self.connect_button = QPushButton("Connect Two Miners", self)
            self.connect_button.clicked.connect(self.toConnectDialog)

            self.newnode_button = QPushButton("Add New Miner", self)
            self.newnode_button.clicked.connect(self.toNewnodeDialog)
    
    def declareBasicButtons(self):
        self.settings_button = QPushButton("Settings", self)
        self.settings_button.clicked.connect(self.toSettings)

        self.flush_temp_button = QPushButton("Flush Temp", self)
        self.flush_temp_button.clicked.connect(self.doFlushTemp)

        self.back_button = QPushButton("Back", self)
        self.back_button.clicked.connect(self.toIdentification)
    
    def setIdentificationButtonsLayoutWidget(self):
        self.identification_layout = QVBoxLayout()
        self.identification_layout.addWidget(self.user_button)
        self.identification_layout.addWidget(self.admin_button)

        self.identification_widget = QWidget()
        self.identification_widget.setFixedSize(800, 400)
        self.identification_widget.setLayout(self.identification_layout)

    def setServiceButtonsLayoutWidget(self, is_admin: bool):
        self.service_layout = QVBoxLayout()
        self.service_layout.addWidget(self.do_query_button)
        self.service_layout.addWidget(self.tx_button)
        if is_admin:
            self.service_layout.addWidget(self.adjust_button)
            self.service_layout.addWidget(self.connect_button)
            self.service_layout.addWidget(self.newnode_button)

        self.service_widget = QWidget()
        self.service_widget.setFixedSize(800, 1000 if is_admin else 400)
        self.service_widget.setLayout(self.service_layout)
    
    def setBasicButtonsLayoutWidget(self):
        self.basic_layout = QHBoxLayout()
        self.basic_layout.addWidget(self.settings_button)
        self.basic_layout.addWidget(self.flush_temp_button)
        self.basic_layout.addWidget(self.back_button)

        self.basic_widget = QWidget()
        self.basic_widget.setFixedSize(1000, 200)
        self.basic_widget.setLayout(self.basic_layout)

    def toUser(self):
        self.clearAllWidgets()

        self.declareServiceButtons(False)
        self.declareBasicButtons()

        self.setServiceButtonsLayoutWidget(False)
        self.setBasicButtonsLayoutWidget()

        self.layout.addWidget(self.service_widget, alignment = Qt.AlignCenter)
        self.layout.addWidget(self.basic_widget, alignment = Qt.AlignBottom)

    def toAdmin(self):
        self.clearAllWidgets()

        self.declareServiceButtons(True)
        self.declareBasicButtons()

        self.setServiceButtonsLayoutWidget(True)
        self.setBasicButtonsLayoutWidget()

        self.layout.addWidget(self.service_widget, alignment = Qt.AlignCenter)
        self.layout.addWidget(self.basic_widget, alignment = Qt.AlignBottom)

    def doQuery(self):
        cmd = f"ssh -p {client_port} {username}@{public_ip} \" \
            cd /osdata/osgroup6/{deployed_repo_name}/; \
            build/helper --command=query \
        \""
        try:
            result = subprocess.run(cmd, shell=True, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
            query_result_dialog = QueryResultDialog(result.stdout.strip().split(" "))
            query_result_dialog.exec_()
        except subprocess.CalledProcessError as e:
            QMessageBox.critical(self, f"Querying Error", f"An error occurred: {e.stderr.strip()}")

    def toTxDialog(self):
        tx_dialog = TxDialog()
        tx_dialog.exec_()

    def toAdjustDialog(self):
        adjust_dialog = AdjustDialog()
        adjust_dialog.exec_()

    def toConnectDialog(self):
        connect_dialog = ConnectDialog()
        connect_dialog.exec_()

    def toNewnodeDialog(self):
        newnode_dialog = NewnodeDialog()
        newnode_dialog.exec_()
    
    def toSettings(self):
        settings_dialog = SettingsDialog()
        settings_dialog.exec_()

    def doFlushTemp(self):
        reply = QMessageBox.question(self, "Flushing Temp Tips",
            "This will cause error if client is on. Do you want to continue?",
            QMessageBox.Yes | QMessageBox.No, QMessageBox.No)
        if reply == QMessageBox.No:
            return
        
        cmd = f"ssh -p {client_port} {username}@{public_ip} \" \
            cd /osdata/osgroup6/{deployed_repo_name}/; \
            build/helper --flush-temp \
        \""
        try:
            result = subprocess.run(cmd, shell=True, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
            QMessageBox.information(self, "Flushing Temp Result", result.stdout.strip())
        except subprocess.CalledProcessError as e:
            QMessageBox.critical(self, f"Flushing Temp Error", f"An error occurred: {e.stderr.strip()}")

    def toIdentification(self):
        self.clearAllWidgets()

        self.declareIdentificationButtons()

        self.setIdentificationButtonsLayoutWidget()

        self.layout.addWidget(self.identification_widget, alignment = Qt.AlignCenter)

 
if __name__ == "__main__":
    app = QApplication(sys.argv)

    # app.setFont(QFont("SimSum", 20))
    app.setFont(QFont("Arial", 20))
    # app.setFont(QFont("Times New Roman", 20))
    # app.setFont(QFont("Consolas", 20))
    
    ex = BitcoinGUI()
    ex.show()
    sys.exit(app.exec_())
