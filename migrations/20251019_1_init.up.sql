CREATE TABLE INTERFACE (
    id SERIAL PRIMARY KEY,
    privateKey VARCHAR NOT NULL,
    address VARCHAR(18) NOT NULL,
    listenPort INTEGER NOT NULL
);

CREATE TABLE PEER (
    id SERIAL PRIMARY KEY,
    interface_id INTEGER REFERENCES INTERFACE (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    publicKey TEXT NOT NULL,
    allowedIPs VARCHAR(18) NOT NULL
);

CREATE TABLE USERS (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE networks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    cidr CIDR NOT NULL UNIQUE,  -- e.g., '192.168.1.0/27'
    owner_id INT REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE network_members (
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    network_id INT REFERENCES networks(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, network_id)
);

CREATE TABLE ip_assignments (
    network_id INT REFERENCES networks(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    ip INET NOT NULL,
    PRIMARY KEY (network_id, ip),
    UNIQUE (user_id, network_id)
);




