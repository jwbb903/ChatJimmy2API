# ChatJimmy Wrapper Go - Python 示例
# 需要先安装：pip install openai

from openai import OpenAI

client = OpenAI(
    base_url="http://127.0.0.1:8787/v1",
    api_key="local-wrapper-key"
)

def main():
    print("=== 非流式聊天 ===")
    response = client.chat.completions.create(
        model="llama3.1-8B",
        messages=[
            {"role": "system", "content": "你是一个有帮助的助手。"},
            {"role": "user", "content": "用一句话介绍你自己。"}
        ]
    )
    print(response.choices[0].message.content)

    print("\n=== 流式聊天 ===")
    stream = client.chat.completions.create(
        model="llama3.1-8B",
        stream=True,
        messages=[
            {"role": "user", "content": "从 1 数到 5。"}
        ]
    )
    for chunk in stream:
        if chunk.choices[0].delta.content:
            print(chunk.choices[0].delta.content, end="", flush=True)
    print()

    print("\n=== 获取模型列表 ===")
    models = client.models.list()
    print(", ".join(m.id for m in models.data))

if __name__ == "__main__":
    main()
