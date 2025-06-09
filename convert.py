import os

def convert_to_lf(filepath):
    """将文件内容中的CRLF替换为LF"""
    with open(filepath, 'rb') as f:
        content = f.read()
    # 替换CRLF为LF
    content = content.replace(b'\r\n', b'\n')
    with open(filepath, 'wb') as f:
        f.write(content)

def process_directory(directory='.'):
    """递归处理目录下的所有文件"""
    for root, _, files in os.walk(directory):
        for filename in files:
            filepath = os.path.join(root, filename)
            try:
                convert_to_lf(filepath)
                print(f"已处理: {filepath}")
            except Exception as e:
                print(f"跳过非文本文件或处理失败: {filepath} - {str(e)}")

if __name__ == '__main__':
    print("开始转换当前目录及子目录下的所有文件...")
    process_directory()
    print("处理完成！")
