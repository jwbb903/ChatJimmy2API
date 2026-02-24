// ChatJimmy Wrapper Go - Node.js 示例
// 需要先安装：npm install openai

import OpenAI from 'openai';

const openai = new OpenAI({
  baseURL: 'http://127.0.0.1:8787/v1',
  apiKey: 'local-wrapper-key',
});

async function main() {
  console.log('=== 非流式聊天 ===');
  const completion = await openai.chat.completions.create({
    model: 'llama3.1-8B',
    messages: [
      { role: 'system', content: '你是一个有帮助的助手。' },
      { role: 'user', content: '用一句话介绍你自己。' },
    ],
  });
  console.log(completion.choices[0].message.content);

  console.log('\n=== 流式聊天 ===');
  const stream = await openai.chat.completions.create({
    model: 'llama3.1-8B',
    stream: true,
    messages: [
      { role: 'user', content: '从 1 数到 5。' },
    ],
  });

  for await (const chunk of stream) {
    const content = chunk.choices[0]?.delta?.content || '';
    if (content) {
      process.stdout.write(content);
    }
  }
  console.log();

  console.log('\n=== 获取模型列表 ===');
  const models = await openai.models.list();
  console.log(models.data.map(m => m.id).join(', '));
}

main().catch(console.error);
