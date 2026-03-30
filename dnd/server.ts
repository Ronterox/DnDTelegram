import express, { Request, Response } from "express";
import cors from "cors";
import { createOpencodeClient } from "@opencode-ai/sdk";
import * as fs from "fs";
import type { OpencodeClient, JsonSchemaFormat, ChatRequestBody, InitRequestBody, ChatResponse, InitResponse, ErrorResponse } from "./types";

const app = express();
app.use(cors());
app.use(express.json({ limit: "50mb" }));

const client: OpencodeClient = createOpencodeClient({
  baseUrl: "http://localhost:4096",
  throwOnError: true,
});

const sessions = new Map<string, boolean>();

interface DndMemory {
  System?: string;
}

async function initSession(sessionId: string | null = null): Promise<string> {
  const memoryContent: string = fs.readFileSync("./dnd.md", "utf-8");
  const memory: DndMemory = JSON.parse(memoryContent);
  
  let id: string = sessionId ?? "";
  
  if (!id) {
    const session = await client.session.create({});
    if (!session.data) {
      throw new Error("Failed to create session");
    }
    id = session.data.id;
    console.log(`Session created: ${id}`);
  }
  
  if (!sessions.has(id)) {
    await client.session.prompt({
      path: { id },
      body: {
        noReply: true,
        parts: [{ type: "text", text: memory.System ?? "" }],
      },
    });
    console.log(`Session initialized with system prompt: ${id}`);
  }
  
  sessions.set(id, true);
  return id;
}

app.post("/api/init", async (req: Request<Record<string, never>, InitResponse | ErrorResponse, InitRequestBody>, res: Response<InitResponse | ErrorResponse>) => {
  try {
    const { sessionId } = req.body;
    const id: string = await initSession(sessionId ?? null);
    res.json({ success: true, sessionId: id });
  } catch (error) {
    const err = error as Error;
    console.error("Init error:", err.message);
    res.status(500).json({ error: err.message });
  }
});

app.post("/api/chat", async (req: Request<Record<string, never>, ChatResponse | ErrorResponse, ChatRequestBody>, res: Response<ChatResponse | ErrorResponse>) => {
  try {
    const { message, format, sessionId } = req.body;
    
    if (!message) {
      return res.status(400).json({ error: "Message is required" });
    }
    
    const id: string = await initSession(sessionId ?? null);
    
    console.log("Sending message to DM:", message.substring(0, 50) + "...");
    
    const body: { parts: Array<{ type: "text"; text: string }>; format?: JsonSchemaFormat } = {
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
    
    const resultData = result.data as { info?: { structured?: Record<string, unknown> }; parts?: Array<{ type: string; text?: string }> } | undefined;
    const structured = resultData?.info?.structured;
    
    if (structured) {
      res.json({ 
        response: structured,
        sessionId: id,
        type: "structured"
      });
    } else {
      const parts = resultData?.parts ?? [];
      const responseText: string = parts
        .filter((p): p is { type: "text"; text: string } => p.type === "text" && typeof p.text === "string")
        .map((p) => p.text)
        .join("\n");
      
      res.json({ 
        response: { narrative: responseText },
        sessionId: id,
        type: "fallback"
      });
    }
  } catch (error) {
    const err = error as Error;
    console.error("Chat error:", err.message);
    res.status(500).json({ error: err.message });
  }
});

app.get("/api/session/:id/messages", async (req: Request<{ id: string }>, res: Response) => {
  try {
    const messages = await client.session.messages({
      path: { id: req.params.id },
    });
    res.json(messages);
  } catch (error) {
    const err = error as Error;
    res.status(500).json({ error: err.message });
  }
});

app.get("/api/health", async (_req: Request, res: Response) => {
  try {
    const response = await fetch("http://localhost:4096/health");
    const health = await response.json();
    res.json(health);
  } catch {
    res.json({ status: "ok", opencode: "connecting" });
  }
});

const PORT: number = parseInt(process.env.PORT ?? "3000", 10);
app.listen(PORT, () => {
  console.log(`API running on http://localhost:${PORT}`);
});
