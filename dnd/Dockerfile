FROM oven/bun:alpine

WORKDIR /app

COPY package.json bun.lock* ./
RUN bun install

COPY . .

EXPOSE 3000 4096

# The command will be overridden in docker-compose for each service
CMD ["bun", "run", "server.ts"]
