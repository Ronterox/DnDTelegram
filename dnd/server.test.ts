import { describe, it, expect, beforeAll, afterAll, mock } from "bun:test";
import express, { Express, Request, Response } from "express";
import cors from "cors";
import type { JsonSchemaFormat } from "./types";

interface MockSessionData {
  id: string;
  parts: Array<{ type: string; text?: string }>;
  info?: { structured: Record<string, unknown> | null };
}

interface ChatRequestBody {
  message?: string;
  sessionId?: string;
  format?: JsonSchemaFormat;
}

interface InitRequestBody {
  sessionId?: string;
}

const mockSession: {
  create: () => Promise<{ data: { id: string } }>;
  prompt: (options: { path: { id: string }; body: unknown }) => Promise<{ data: MockSessionData }>;
  messages: (options: { path: { id: string } }) => Promise<{ data: unknown[] }>;
} = {
  create: async () => ({ data: { id: "test-session-id" } }),
  prompt: async () => ({
    data: {
      id: "msg-id",
      parts: [{ type: "text", text: "Mock DM response" }],
      info: { structured: null }
    }
  }),
  messages: async () => ({ data: [] })
};

mock.module("@opencode-ai/sdk", () => ({
  createOpencodeClient: () => ({ session: mockSession }) as never
}));

const app: Express = express();
app.use(cors());
app.use(express.json({ limit: "50mb" }));

const sessions = new Map<string, boolean>();
let client: { session: typeof mockSession };
let server: ReturnType<typeof app.listen>;

beforeAll(async () => {
  const { createOpencodeClient } = await import("@opencode-ai/sdk");
  
  client = createOpencodeClient({
    baseUrl: "http://localhost:4096",
    throwOnError: true,
  }) as unknown as { session: typeof mockSession };

  async function initSession(sessionId: string | null = null): Promise<string> {
    const memory = { System: "Test DM memory" };
    let id: string = sessionId ?? "";
    if (!id) {
      const session = await client.session.create();
      if (!session.data) {
        throw new Error("Failed to create session");
      }
      id = session.data.id;
    }
    if (!sessions.has(id)) {
      await client.session.prompt({
        path: { id },
        body: { noReply: true, parts: [{ type: "text", text: memory.System }] },
      });
    }
    sessions.set(id, true);
    return id;
  }

  app.post("/api/init", async (req: Request<Record<string, never>, unknown, InitRequestBody>, res: Response) => {
    try {
      const { sessionId } = req.body as InitRequestBody;
      const id: string = await initSession(sessionId ?? null);
      res.json({ success: true, sessionId: id });
    } catch (error) {
      const err = error as Error;
      res.status(500).json({ error: err.message });
    }
  });

  app.post("/api/chat", async (req: Request<Record<string, never>, unknown, ChatRequestBody>, res: Response) => {
    try {
      const body = req.body as ChatRequestBody;
      const { message, format, sessionId } = body;
      if (!message) {
        res.status(400).json({ error: "Message is required" });
        return;
      }
      const id: string = await initSession(sessionId ?? null);
      const promptBody: { parts: Array<{ type: "text"; text: string }>; format?: JsonSchemaFormat } = { parts: [{ type: "text", text: message }] };
      if (format) promptBody.format = format;
      const result = await client.session.prompt({ path: { id }, body: promptBody });
      const data = result.data as MockSessionData | undefined;
      const structured = data?.info?.structured;
      if (structured) {
        res.json({ response: structured, sessionId: id, type: "structured" });
      } else {
        const parts = data?.parts ?? [];
        const responseText: string = parts.filter((p): p is { type: "text"; text: string } => p.type === "text" && typeof p.text === "string").map((p) => p.text).join("\n");
        res.json({ response: { narrative: responseText }, sessionId: id, type: "fallback" });
      }
    } catch (error) {
      const err = error as Error;
      res.status(500).json({ error: err.message });
    }
  });

  app.get("/api/session/:id/messages", async (req: Request<{ id: string }>, res: Response) => {
    try {
      const messages = await client.session.messages({ path: { id: req.params.id } });
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

  server = app.listen(3000);
});

afterAll(() => {
  server?.close();
});

describe("POST /api/init", () => {
  it("creates a new session when no sessionId provided", async () => {
    const res = await fetch("http://localhost:3000/api/init", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({})
    });
    const data = await res.json() as { success: boolean; sessionId: string };
    expect(data.success).toBe(true);
    expect(data.sessionId).toBe("test-session-id");
  });

  it("uses provided sessionId if given", async () => {
    const res = await fetch("http://localhost:3000/api/init", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ sessionId: "existing-id" })
    });
    const data = await res.json() as { sessionId: string };
    expect(data.sessionId).toBe("existing-id");
  });
});

describe("POST /api/chat", () => {
  it("returns 400 if message is missing", async () => {
    const res = await fetch("http://localhost:3000/api/chat", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({})
    });
    const data = await res.json() as { error: string };
    expect(res.status).toBe(400);
    expect(data.error).toBe("Message is required");
  });

  it("returns DM response with narrative", async () => {
    const res = await fetch("http://localhost:3000/api/chat", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ message: "Hello DM" })
    });
    const data = await res.json() as { sessionId: string; response: { narrative: string }; type: string };
    expect(data.sessionId).toBeDefined();
    expect(data.response.narrative).toBe("Mock DM response");
    expect(data.type).toBe("fallback");
  });

  it("returns structured response when format provided", async () => {
    const format: JsonSchemaFormat = {
      type: "json_schema",
      schema: {
        type: "object",
        properties: {
          narrative: { type: "string" },
          diceRoll: { type: "object", properties: { type: { type: "string" } } }
        },
        required: ["narrative"]
      }
    };
    mockSession.prompt = async () => ({
      data: {
        id: "msg-id",
        parts: [],
        info: { structured: { narrative: "Structured response", diceRoll: { type: "strength" } } }
      }
    });
    const res = await fetch("http://localhost:3000/api/chat", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ message: "Roll strength", format })
    });
    const data = await res.json() as { type: string; response: { narrative: string; diceRoll: { type: string } } };
    expect(data.type).toBe("structured");
    expect(data.response.narrative).toBe("Structured response");
    expect(data.response.diceRoll).toEqual({ type: "strength" });
  });
});

describe("GET /api/session/:id/messages", () => {
  it("returns messages for a session", async () => {
    const res = await fetch("http://localhost:3000/api/session/test-id/messages");
    const data = await res.json() as { data: unknown[] };
    expect(data.data).toEqual([]);
  });
});

describe("GET /api/health", () => {
  it("returns health status", async () => {
    const res = await fetch("http://localhost:3000/api/health");
    const data = await res.json() as { status: string; opencode: string };
    expect(data.status).toBe("ok");
    expect(data.opencode).toBe("connecting");
  });
});
