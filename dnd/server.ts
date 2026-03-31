import { createOpencodeClient } from "@opencode-ai/sdk";
import cors from "cors";
import crypto from "crypto";
import express, { Request, Response } from "express";
import * as fs from "fs";
import type {
    ChatRequestBody,
    ChatResponse,
    ErrorResponse,
    InitRequestBody,
    InitResponse,
    JsonSchemaFormat,
    OpencodeClient,
} from "./types";

import jwt, { JwtPayload } from "jsonwebtoken";

export const app = express();
app.use(cors());
app.use(express.json({ limit: "50mb" }));

let client: OpencodeClient = createOpencodeClient({
    baseUrl: "http://localhost:4096",
    throwOnError: true,
});

export function setClient(mockClient: OpencodeClient): void {
    client = mockClient;
}

export const sessions = new Map<string, boolean>();

let healthCheckOverride: (() => Promise<unknown>) | null = null;

export function setHealthCheckOverride(fn: () => Promise<unknown>): void {
    healthCheckOverride = fn;
}

export async function initSession(
    sessionId: string | null = null,
): Promise<string> {
    const memoryContent: string = fs.readFileSync("./dnd.md", "utf-8");
    console.log("Loaded dnd.md, length:", memoryContent.length);

    let id: string = sessionId ?? "";

    if (!id) {
        console.log("Creating new session...");
        console.log("Checking OpenCode server at http://localhost:4096...");
        let session;
        try {
            const healthCheck = await fetch("http://localhost:4096/health");
            console.log(
                "Health check response:",
                healthCheck.status,
                await healthCheck.text(),
            );
            session = await client.session.create({});
        } catch (e) {
            console.error("session.create threw:", e);
            throw e;
        }
        console.log("Session create response:", JSON.stringify(session));
        if (!session.data) {
            throw new Error(
                "Failed to create session: " + JSON.stringify(session),
            );
        }
        id = session.data.id;
        console.log(`Session created: ${id}`);
    }

    if (!sessions.has(id)) {
        console.log("Prompting session with memory...");
        await client.session.prompt({
            path: { id },
            body: {
                noReply: true,
                parts: [{ type: "text", text: memoryContent }],
            },
        });
        console.log(`Session initialized with system prompt: ${id}`);
    }

    sessions.set(id, true);
    return id;
}

app.post(
    "/api/init",
    async (
        req: Request<
            Record<string, never>,
            InitResponse | ErrorResponse,
            InitRequestBody
        >,
        res: Response<InitResponse | ErrorResponse>,
    ) => {
        try {
            const { sessionId } = req.body;
            const id: string = await initSession(sessionId ?? null);
            res.json({ success: true, sessionId: id });
        } catch (error) {
            const err = error as Error;
            console.error("Init error:", error);
            console.error("Init error type:", typeof error);
            if (error instanceof Error) {
                console.error("Init error:", err.message, err.stack);
                res.status(500).json({ error: err.message || String(error) });
            } else {
                res.status(500).json({ error: String(error) });
            }
        }
    },
);

app.post(
    "/api/chat",
    async (
        req: Request<
            Record<string, never>,
            ChatResponse | ErrorResponse,
            ChatRequestBody
        >,
        res: Response<ChatResponse | ErrorResponse>,
    ) => {
        try {
            const { message, format, sessionId } = req.body;

            if (!message) {
                return res.status(400).json({ error: "Message is required" });
            }

            const id: string = await initSession(sessionId ?? null);

            console.log(
                "Sending message to DM:",
                message.substring(0, 50) + "...",
            );

            const body: {
                parts: Array<{ type: "text"; text: string }>;
                format?: JsonSchemaFormat;
            } = {
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

            const resultData = result.data as
                | {
                      info?: { structured?: Record<string, unknown> };
                      parts?: Array<{ type: string; text?: string }>;
                  }
                | undefined;
            const structured = resultData?.info?.structured;

            if (structured) {
                res.json({
                    response: structured,
                    sessionId: id,
                    type: "structured",
                });
            } else {
                const parts = resultData?.parts ?? [];
                const responseText: string = parts
                    .filter(
                        (p): p is { type: "text"; text: string } =>
                            p.type === "text" && typeof p.text === "string",
                    )
                    .map((p) => p.text)
                    .join("\n");

                res.json({
                    response: { narrative: responseText },
                    sessionId: id,
                    type: "fallback",
                });
            }
        } catch (error) {
            const err = error as Error;
            console.error("Chat error:", err.message);
            res.status(500).json({ error: err.message });
        }
    },
);

app.get(
    "/api/session/:id/messages",
    async (req: Request<{ id: string }>, res: Response) => {
        try {
            const messages = await client.session.messages({
                path: { id: req.params.id },
            });
            res.json(messages);
        } catch (error) {
            const err = error as Error;
            res.status(500).json({ error: err.message });
        }
    },
);

app.get("/api/health", async (_req: Request, res: Response) => {
    try {
        if (healthCheckOverride) {
            const health = await healthCheckOverride();
            res.json(health);
            return;
        }
        const response = await fetch("http://localhost:4096/health");
        const health = await response.json();
        res.json(health);
    } catch {
        res.json({ status: "ok", opencode: "connecting" });
    }
});

app.get("/api/auth/login", async (req: Request, res: Response) => {
    const { email, password } = req.body;
    if (!email || !password) {
        return res
            .status(400)
            .json({ error: "email and password are required" });
    }

    const hashedPassword = crypto
        .createHash("sha256")
        .update(password)
        .digest("hex");
    if (hashedPassword !== process.env.PASSWORD) {
        return res.status(401).json({ error: "UNAUTHORIZED" });
    }
    const user = { email, username: email };
    const token = encodeUserToken(user);
    res.json({ token });
});

app.post("/api/auth/register", async (req: Request, res: Response) => {
    const { username, email, password } = req.body;
    if (!username || !email || !password) {
        return res
            .status(400)
            .json({ error: "username, email and password are required" });
    }

    const userExists = false;
    if (userExists)
        return res.status(400).json({ error: "User already exists" });

    const hashedPassword = crypto
        .createHash("sha256")
        .update(password)
        .digest("hex");

    const newUser = { email, username, password: hashedPassword };
    const token = encodeUserToken(newUser);

    res.json({ token });
});

app.post("/api/auth/authenticate", async (req: Request, res: Response) => {
    const { token } = req.body;

    const decoded = decodeUserToken(token);
    if (!decoded) {
        return res.status(401).json({ error: "Invalid token" });
    }
    res.json({ success: true, user: decoded });
});

function encodeUserToken(user: { email: string; username: string }): string {
    return jwt.sign(user, process.env.AUTH_SECRET!, {
        expiresIn: "1h",
    });
}

function decodeUserToken(
    token: string,
): { email: string; username: string } | null {
    const decoded = jwt.verify(token, process.env.AUTH_SECRET!) as JwtPayload;
    if (!decoded || decoded.exp! < Date.now() / 1000) {
        return null;
    }
    return decoded as { email: string; username: string };
}

export function startServer(
    port: number = 3000,
): ReturnType<typeof app.listen> {
    return app.listen(port, () => {
        console.log(`API running on http://localhost:${port}`);
    });
}

if (process.env.NODE_ENV !== "test") {
    startServer();
}
