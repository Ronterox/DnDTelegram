import express from 'express';
import cors from 'cors';
import Redis from 'ioredis';

const app = express();
app.use(cors());
app.use(express.json({ limit: '50mb' }));

const redis = new Redis({
  host: '127.0.0.1',
  port: 6379,
});

redis.on('connect', () => {
  console.log('Connected to Redis');
});

redis.on('error', (err) => {
  console.error('Redis error:', err);
});

const GAME_KEY_PREFIX = 'game:';

async function addToIndex(sessionId, gameData) {
  for (const [key, value] of Object.entries(gameData)) {
    if (value !== null && typeof value !== 'object' && key !== 'session_id') {
      const indexKey = `game:index:${key}:${value}`;
      await redis.sadd(indexKey, sessionId);
    }
    
    if (key === 'players' && Array.isArray(value)) {
      for (const player of value) {
        if (player.name) {
          const nameIndexKey = `game:index:player:name:${player.name.toLowerCase()}`;
          await redis.sadd(nameIndexKey, sessionId);
        }
      }
    }
  }
  
  await redis.sadd('game:all_ids', sessionId);
}

async function removeFromIndex(sessionId) {
  await redis.srem('game:all_ids', sessionId);
  
  const keys = await redis.keys(`game:index:*:${sessionId}`);
  if (keys.length > 0) {
    await redis.del(...keys);
  }
}

app.post('/api/games', async (req, res) => {
  try {
    const gameData = req.body;
    
    if (!gameData.session_id) {
      return res.status(400).json({ error: 'session_id is required' });
    }

    const sessionId = gameData.session_id;
    const key = `${GAME_KEY_PREFIX}${sessionId}`;
    
    await redis.set(key, JSON.stringify(gameData));
    await addToIndex(sessionId, gameData);
    
    res.status(201).json({ message: 'Game saved', session_id: sessionId });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.get('/api/games/:sessionId', async (req, res) => {
  try {
    const { sessionId } = req.params;
    const key = `${GAME_KEY_PREFIX}${sessionId}`;
    
    const game = await redis.get(key);
    
    if (!game) {
      return res.status(404).json({ error: 'Game not found' });
    }
    
    res.json(JSON.parse(game));
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.get('/api/games', async (req, res) => {
  try {
    const { field, value } = req.query;
    
    if (field && value) {
      let sessionIds;
      
      if (field === 'player_name') {
        const indexKey = `game:index:player:name:${value.toLowerCase()}`;
        sessionIds = await redis.smembers(indexKey);
      } else {
        const indexKey = `game:index:${field}:${value}`;
        sessionIds = await redis.smembers(indexKey);
      }
      
      if (!sessionIds || sessionIds.length === 0) {
        return res.json([]);
      }
      
      const pipeline = redis.pipeline();
      for (const sid of sessionIds) {
        pipeline.get(`${GAME_KEY_PREFIX}${sid}`);
      }
      
      const results = await pipeline.exec();
      const games = results
        .map(([err, data]) => data ? JSON.parse(data) : null)
        .filter(Boolean);
      
      return res.json(games);
    }
    
    const allIds = await redis.smembers('game:all_ids');
    
    if (!allIds || allIds.length === 0) {
      return res.json([]);
    }
    
    const pipeline = redis.pipeline();
    for (const sid of allIds) {
      pipeline.get(`${GAME_KEY_PREFIX}${sid}`);
    }
    
    const results = await pipeline.exec();
    const games = results
      .map(([err, data]) => data ? JSON.parse(data) : null)
      .filter(Boolean);
    
    res.json(games);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.delete('/api/games/:sessionId', async (req, res) => {
  try {
    const { sessionId } = req.params;
    const key = `${GAME_KEY_PREFIX}${sessionId}`;
    
    const exists = await redis.exists(key);
    if (!exists) {
      return res.status(404).json({ error: 'Game not found' });
    }
    
    await redis.del(key);
    await removeFromIndex(sessionId);
    
    res.json({ message: 'Game deleted' });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.put('/api/games/:sessionId', async (req, res) => {
  try {
    const { sessionId } = req.params;
    const gameData = req.body;
    const key = `${GAME_KEY_PREFIX}${sessionId}`;
    
    const exists = await redis.exists(key);
    if (!exists) {
      return res.status(404).json({ error: 'Game not found' });
    }
    
    gameData.session_id = sessionId;
    await redis.set(key, JSON.stringify(gameData));
    await removeFromIndex(sessionId);
    await addToIndex(sessionId, gameData);
    
    res.json({ message: 'Game updated', session_id: sessionId });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

const PORT = process.env.PORT || 3001;
app.listen(PORT, () => {
  console.log(`Game API running on http://localhost:${PORT}`);
});
