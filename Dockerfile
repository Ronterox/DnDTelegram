FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    curl \
    git \
    wget \
    ca-certificates \
    gnupg \
    lsb-release \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

RUN apt-get update && apt-get install -y redis-server golang-go \
    && rm -rf /var/lib/apt/lists/*

RUN wget https://github.com/oven-sh/bun/releases/latest/download/bun-linux-x64.zip -O /tmp/bun.zip \
    && unzip /tmp/bun.zip -d /usr/local/bin \
    && rm /tmp/bun.zip \
    && chmod +x /usr/local/bin/bun

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . /app

RUN cd /app/dnd && bun install
RUN cd /app/SixSevenStory && bun install

ENV TOKEN=""
ENV REDIS_HOST=localhost

EXPOSE 3000 3001 4096 5173

CMD redis-server --daemonize yes && \
    (cd /app/dnd && bun run opencode) & \
    (cd /app/dnd && bun start) & \
    (cd /app && bun run game-api.js) & \
    (cd /app/SixSevenStory && bun run dev) & \
    (sleep 8 && go run /app/main.go) & \
    wait