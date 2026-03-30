import { describe, it, expect, beforeAll, beforeEach, afterAll, mock } from 'bun:test';
import express from 'express';
import cors from 'cors';

const mockSession = {
  create: async () => ({ data: { id: 'test-session-id' } }),
  prompt: async () => ({
    data: {
      parts: [{ type: 'text', text: 'Mock DM response' }],
      info: { structured: null }
    }
  }),
  messages: async () => ({ data: [] })
};

mock.module('@opencode-ai/sdk', () => ({
  createOpencodeClient: () => ({ session: mockSession })
}));

const app = express();
app.use(cors());
app.use(express.json({ limit: '50mb' }));

const sessions = new Map();
let client;
let server;

beforeAll(async () => {
  const { createOpencodeClient } = await import('@opencode-ai/sdk');
  
  client = createOpencodeClient({
    baseUrl: 'http://localhost:4096',
    throwOnError: true,
  });

  async function initSession(sessionId = null) {
    const memory = { System: 'Test DM memory' };
    let id = sessionId;
    if (!id) {
      const session = await client.session.create({
        body: { title: 'Dungeons & Dragons - Dungeon Master' },
      });
      id = session.data.id;
    }
    if (!sessions.has(id)) {
      await client.session.prompt({
        path: { id },
        body: { noReply: true, parts: [{ type: 'text', text: memory.System }] },
      });
    }
    sessions.set(id, true);
    return id;
  }

  app.post('/api/init', async (req, res) => {
    try {
      const { sessionId } = req.body;
      const id = await initSession(sessionId);
      res.json({ success: true, sessionId: id });
    } catch (error) {
      res.status(500).json({ error: error.message });
    }
  });

  app.post('/api/chat', async (req, res) => {
    try {
      const { message, format, sessionId } = req.body;
      if (!message) {
        return res.status(400).json({ error: 'Message is required' });
      }
      const id = await initSession(sessionId);
      const body = { parts: [{ type: 'text', text: message }] };
      if (format) body.format = format;
      const result = await client.session.prompt({ path: { id }, body });
      const structured = result.data.info?.structured;
      if (structured) {
        res.json({ response: structured, sessionId: id, type: 'structured' });
      } else {
        const parts = result.data.parts || [];
        const responseText = parts.filter(p => p.type === 'text').map(p => p.text).join('\n');
        res.json({ response: { narrative: responseText }, sessionId: id, type: 'fallback' });
      }
    } catch (error) {
      res.status(500).json({ error: error.message });
    }
  });

  app.get('/api/session/:id/messages', async (req, res) => {
    try {
      const messages = await client.session.messages({ path: { id: req.params.id } });
      res.json(messages);
    } catch (error) {
      res.status(500).json({ error: error.message });
    }
  });

  app.get('/api/health', async (req, res) => {
    try {
      const response = await fetch('http://localhost:4096/health');
      const health = await response.json();
      res.json(health);
    } catch (error) {
      res.json({ status: 'ok', opencode: 'connecting' });
    }
  });

  server = app.listen(3000);
});

afterAll(() => {
  server?.close();
});

describe('POST /api/init', () => {
  it('creates a new session when no sessionId provided', async () => {
    const res = await fetch('http://localhost:3000/api/init', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({})
    });
    const data = await res.json();
    expect(data.success).toBe(true);
    expect(data.sessionId).toBe('test-session-id');
  });

  it('uses provided sessionId if given', async () => {
    const res = await fetch('http://localhost:3000/api/init', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ sessionId: 'existing-id' })
    });
    const data = await res.json();
    expect(data.sessionId).toBe('existing-id');
  });
});

describe('POST /api/chat', () => {
  it('returns 400 if message is missing', async () => {
    const res = await fetch('http://localhost:3000/api/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({})
    });
    const data = await res.json();
    expect(res.status).toBe(400);
    expect(data.error).toBe('Message is required');
  });

  it('returns DM response with narrative', async () => {
    const res = await fetch('http://localhost:3000/api/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message: 'Hello DM' })
    });
    const data = await res.json();
    expect(data.sessionId).toBeDefined();
    expect(data.response.narrative).toBe('Mock DM response');
    expect(data.type).toBe('fallback');
  });

  it('returns structured response when format provided', async () => {
    const format = {
      type: 'json_schema',
      schema: {
        type: 'object',
        properties: {
          narrative: { type: 'string' },
          diceRoll: { type: 'object', properties: { type: { type: 'string' } } }
        },
        required: ['narrative']
      }
    };
    mockSession.prompt = async () => ({
      data: {
        parts: [],
        info: { structured: { narrative: 'Structured response', diceRoll: { type: 'strength' } } }
      }
    });
    const res = await fetch('http://localhost:3000/api/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message: 'Roll strength', format })
    });
    const data = await res.json();
    expect(data.type).toBe('structured');
    expect(data.response.narrative).toBe('Structured response');
    expect(data.response.diceRoll).toEqual({ type: 'strength' });
  });
});

describe('GET /api/session/:id/messages', () => {
  it('returns messages for a session', async () => {
    const res = await fetch('http://localhost:3000/api/session/test-id/messages');
    const data = await res.json();
    expect(data.data).toEqual([]);
  });
});

describe('GET /api/health', () => {
  it('returns health status', async () => {
    const res = await fetch('http://localhost:3000/api/health');
    const data = await res.json();
    expect(data.status).toBe('ok');
    expect(data.opencode).toBe('connecting');
  });
});
