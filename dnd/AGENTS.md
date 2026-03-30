# D&D API Agent Integration

This API provides a chat interface for a Dungeon Master powered by OpenCode AI.

## Setup

Start both servers:

```bash
# Terminal 1: Start OpenCode server
npm run opencode

# Terminal 2: Start API server
npm start
```

## Endpoints

### `POST /api/chat`

Send a message to the Dungeon Master.

```javascript
fetch('http://localhost:3000/api/chat', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    message: "I want to start a new adventure",
    sessionId: null,  // omit or null to create new session
    format: {         // optional: request structured output
      type: "json_schema",
      schema: {
        type: "object",
        properties: {
          narrative: { type: "string" },
          diceRoll: { 
            type: "object",
            properties: {
              type: { type: "string" },
              dc: { type: "number" },
              result: { type: "string" }
            }
          }
        },
        required: ["narrative"]
      }
    }
  })
})
```

The `format` parameter uses OpenCode's structured output format. Pass a Zod JSON schema to force the AI to respond in a specific structure.

**Response (when format specified):**

```json
{
  "response": {
    "narrative": "The DM's response text...",
    "diceRoll": { "type": "dexterity", "dc": 14, "result": "success" }
  },
  "sessionId": "abc-123",
  "type": "structured"
}
```

**Response (fallback without format):**

```json
{
  "response": {
    "narrative": "The DM's response text..."
  },
  "sessionId": "abc-123",
  "type": "fallback"
}
```

**Response:**

```json
{
  "response": {
    "narrative": "The DM's response text..."
  },
  "sessionId": "abc-123",
  "type": "structured" | "fallback"
}
```

### `POST /api/init`

Initialize a new or existing session.

```javascript
fetch('http://localhost:3000/api/init', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    sessionId: null  // omit to create new session
  })
})
```

**Response:**

```json
{
  "success": true,
  "sessionId": "abc-123"
}
```

### `GET /api/health`

Check API and OpenCode server health.

### `GET /api/session/:id/messages`

Get message history for a session.

## Usage in Agents

For skill integration, make HTTP requests to `/api/chat`:

```javascript
async function chatWithDM(message, sessionId = null) {
  const res = await fetch('http://localhost:3000/api/chat', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ message, sessionId })
  });
  const data = await res.json();
  return {
    response: data.response.narrative,
    sessionId: data.sessionId
  };
}
```

For structured output with Zod JSON schema:

```javascript
const format = {
  type: "json_schema",
  schema: {
    type: "object",
    properties: {
      narrative: { type: "string" },
      diceRoll: {
        type: "object",
        properties: {
          type: { type: "string" },
          dc: { type: "number" },
          result: { type: "string" }
        }
      }
    },
    required: ["narrative"]
  }
};

const res = await fetch('http://localhost:3000/api/chat', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ message, sessionId, format })
});
```

## Session Management

- Pass `sessionId` to continue an existing conversation
- Pass `null` or omit to create a new session
- Sessions persist the DM's memory of the campaign
