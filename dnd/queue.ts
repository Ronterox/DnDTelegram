import amqp from "amqplib";
import { initSession, handleChat } from "./server";

const RABBITMQ_URL = process.env.RABBITMQ_URL || "amqp://guest:guest@rabbitmq:5672";

async function startQueue() {
    let connection;
    for (let i = 0; i < 10; i++) {
        try {
            connection = await amqp.connect(RABBITMQ_URL);
            break;
        } catch (e) {
            console.log(`Failed to connect to RabbitMQ, retrying in 5s... (${i + 1}/10)`);
            await new Promise(resolve => setTimeout(resolve, 5000));
        }
    }

    if (!connection) {
        console.error("Could not connect to RabbitMQ");
        process.exit(1);
    }

    const channel = await connection.createChannel();

    // Setup DLX
    await channel.assertExchange("dlx_exchange", "direct", { durable: true });
    await channel.assertQueue("dlx_queue", { durable: true });
    await channel.bindQueue("dlx_queue", "dlx_exchange", "dlx_routing_key");

    const queue = "rpc_queue";
    await channel.assertQueue(queue, {
        durable: true,
        arguments: {
            "x-dead-letter-exchange": "dlx_exchange",
            "x-dead-letter-routing-key": "dlx_routing_key",
            "x-max-priority": 10
        }
    });

    channel.prefetch(1);
    console.log(" [x] Awaiting RPC requests");

    channel.consume(queue, async (msg) => {
        if (!msg) return;

        const content = JSON.parse(msg.content.toString());
        const { method, payload, sessionId } = content;
        
        console.log(` [.] Received request: ${method}`);

        let response;
        try {
            if (method === "init") {
                const id = await initSession(sessionId || null);
                response = { success: true, sessionId: id };
            } else if (method === "chat") {
                const result = await handleChat(payload.message, sessionId, payload.format);
                response = { success: true, ...result };
            } else {
                response = { success: false, error: "Unknown method" };
            }
        } catch (e) {
            const err = e as Error;
            console.error(`Error processing ${method}:`, err.message);
            response = { success: false, error: err.message };
        }

        channel.sendToQueue(msg.properties.replyTo, Buffer.from(JSON.stringify(response)), {
            correlationId: msg.properties.correlationId
        });

        channel.ack(msg);
    });
}

if (process.env.NODE_ENV !== "test") {
    startQueue().catch(console.error);
}
