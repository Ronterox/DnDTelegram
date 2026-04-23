FROM oven/bun:alpine

WORKDIR /app

# Install dependencies required for the opencode installer
RUN apk add --no-cache curl bash

# Install opencode binary
RUN curl -fsSL https://opencode.ai/install | bash

COPY package.json bun.lock* ./
RUN bun install

COPY . .

EXPOSE 3000 4096

# The command will be overridden in docker-compose for each service
CMD ["bun", "run", "server.ts"]
