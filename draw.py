import matplotlib.pyplot as plt

# 数据点
log_difficulty = [18, 20, 22, 24]  # 横轴数据
log_block_numbers_per_minute = [1.222392421, 3.273018494, 5.2605755, 7.400879436]  # 纵轴数据

# 创建图形
plt.figure(figsize=(8, 6))  # 设置图形大小
plt.scatter(log_difficulty, log_block_numbers_per_minute, color='red', label='Data Points')  # 绘制散点图

plt.plot(log_difficulty, log_block_numbers_per_minute, color='blue', label='Data Points')  # 绘制折线图

# 添加标题和标签
plt.title('Line Plot of log_difficulty vs log_block_numbers_per_minute')
plt.xlabel('log_difficulty')
plt.ylabel('log_block_numbers_per_minute')

# 添加网格
plt.grid(True, linestyle='--', alpha=0.6)

# 显示图例
plt.legend()

# 显示图形
plt.show()

# 保存图形
plt.savefig('img_blockchain/effi_plot.png')