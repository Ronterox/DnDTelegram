import type { OpencodeClient, Session } from "@opencode-ai/sdk";
export type User = {
    email: string;
    username: string;
    password: string;
};

export interface JsonSchemaProperty {
    type: "string" | "number" | "boolean" | "object" | "array";
    properties?: Record<string, JsonSchemaProperty>;
    required?: string[];
}

export interface JsonSchemaFormat {
    type: "json_schema";
    schema: JsonSchemaProperty;
}

export interface ChatRequestBody {
    message: string;
    sessionId?: string | null;
    format?: JsonSchemaFormat;
}

export interface InitRequestBody {
    sessionId?: string | null;
}

export interface ChatResponse {
    response: Record<string, unknown>;
    sessionId: string;
    type: "structured" | "fallback";
}

export interface InitResponse {
    success: boolean;
    sessionId: string;
}

export interface ErrorResponse {
    error: string;
}

export type { OpencodeClient, Session };
