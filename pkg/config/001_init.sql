-- Users
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Teams
CREATE TABLE teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Relación usuarios ↔ teams (muchos a muchos)
CREATE TABLE user_teams (
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    team_id INT REFERENCES teams(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member', -- admin / member
    PRIMARY KEY (user_id, team_id)
);

-- Channels
CREATE TABLE channels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    team_id INT REFERENCES teams(id) ON DELETE CASCADE, -- Nulable para DMs
    is_dm BOOLEAN NOT NULL DEFAULT FALSE, -- True para canales de DM
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Relación usuarios ↔ canales (muchos a muchos con rol)
CREATE TABLE channel_users (
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    channel_id INT REFERENCES channels(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'user', -- admin / user
    PRIMARY KEY (user_id, channel_id)
);

-- Messages
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    channel_id INT REFERENCES channels(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Amistades entre usuarios
CREATE TABLE friends (
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    friend_id INT REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending', -- pending, accepted, blocked
    PRIMARY KEY (user_id, friend_id)
);

-- Last Read
CREATE TABLE last_read (
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    channel_id INT REFERENCES channels(id) ON DELETE CASCADE,
    message_id INT REFERENCES messages(id) ON DELETE CASCADE,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, channel_id)
);
