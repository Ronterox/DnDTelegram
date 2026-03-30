import { describe, it, expect, beforeAll, afterAll, beforeEach } from "bun:test";
import { app, setClient, sessions, setHealthCheckOverride, startServer } from "./server";
import type { JsonSchemaFormat, OpencodeClient } from "./types";

interface MockSessionData {
  id: string;
  parts: Array<{ type: string; text?: string }>;
  info?: { structured: Record<string, unknown> | null };
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

const mockClient = { session: mockSession } as unknown as OpencodeClient;

let server: ReturnType<typeof app.listen>;

beforeAll(() => {
  setClient(mockClient);
  setHealthCheckOverride(async () => ({ status: "ok", opencode: "connected" }));
  server = startServer(3000);
});

afterAll(() => {
  server?.close();
});

beforeEach(() => {
  sessions.clear();
  mockSession.prompt = async () => ({
    data: {
      id: "msg-id",
      parts: [{ type: "text", text: "Mock DM response" }],
      info: { structured: null }
    }
  });
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
    expect(data.opencode).toBe("connected");
  });
});
