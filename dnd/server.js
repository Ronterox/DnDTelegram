import express from 'express';
import cors from 'cors';
import { createOpencodeClient } from "@opencode-ai/sdk"
import fs from 'fs';

const app = express();
app.use(cors());
app.use(express.json({ limit: '50mb' }));

const client = createOpencodeClient({
  baseUrl: "http://localhost:4096",
  throwOnError: true,
});

const sessions = new Map();

async function initSession(sessionId = null) {
  const memory = JSON.parse(fs.readFileSync('./dnd.md', 'utf-8'));
  
  let id = sessionId;
  
  if (!id) {
    const session = await client.session.create({
      body: { title: "Dungeons & Dragons - Dungeon Master" },
    });
    id = session.data.id;
    console.log(`Session created: ${id}`);
  }
  
  if (!sessions.has(id)) {
    await client.session.prompt({
      path: { id },
      body: {
        noReply: true,
        parts: [{ type: "text", text: memory.System }],
      },
    });
    console.log(`Session initialized with system prompt: ${id}`);
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
    console.error("Init error:", error.message);
    res.status(500).json({ error: error.message });
  }
});

app.post('/api/chat', async (req, res) => {
  try {
    const { message, format, sessionId } = req.body;
    
    if (!message) {
      return res.status(400).json({ error: "Message is required" });
    }
    
    const id = await initSession(sessionId);
    
    console.log("Sending message to DM:", message.substring(0, 50) + "...");
    
    const body = {
      parts: [{ type: "text", text: message }],
    };
    
    if (format) {
      body.format = format;
    }
    
    const result = await client.session.prompt({
      path: { id },
      body,
    });
    
    console.log("DM response received");
    
    const structured = result.data.info?.structured;
    
    if (structured) {
      res.json({ 
        response: structured,
        sessionId: id,
        type: "structured"
      });
    } else {
      const parts = result.data.parts || [];
      const responseText = parts
        .filter(p => p.type === "text")
        .map(p => p.text)
        .join("\n");
      
      res.json({ 
        response: { narrative: responseText },
        sessionId: id,
        type: "fallback"
      });
    }
  } catch (error) {
    console.error("Chat error:", error.message);
    res.status(500).json({ error: error.message });
  }
});

app.get('/api/session/:id/messages', async (req, res) => {
  try {
    const messages = await client.session.messages({
      path: { id: req.params.id },
    });
    res.json(messages);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.get('/api/health', async (req, res) => {
  try {
    const response = await fetch("http://localhost:4096/health");
    const health = await response.json();
    res.json(health);
  } catch (error) {
    res.json({ status: "ok", opencode: "connecting" });
  }
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`API running on http://localhost:${PORT}`);
});
