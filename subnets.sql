-- name: CreateUser :one
INSERT INTO users (username, email)
VALUES ($1, $2)
RETURNING id, username, email, created_at;

-- name: CreateNetwork :one
INSERT INTO networks (name, cidr, owner_id)
VALUES ($1, $2, $3)
RETURNING id, name, cidr, owner_id, created_at;

-- name: ListNetworks :many
SELECT id, name, cidr, owner_id, created_at
FROM networks
ORDER BY created_at DESC;


-- name: OwnerHasNetwork :one
SELECT EXISTS (
    SELECT 1 FROM networks WHERE owner_id = $1
) AS has_network;

-- name: GetNetworkByOwner :one
SELECT id, name, cidr, owner_id, created_at
FROM networks
WHERE owner_id = $1
LIMIT 1;


-- name: GetSubnetsWithPeers :many
SELECT i.id AS interface_id, i.address, i.listenPort,
       p.id AS peer_id, p.name, p.publicKey, p.allowedIPs
FROM INTERFACE i
LEFT JOIN PEER p ON p.interface_id = i.id;

-- name: CheckIpInNetwork :one
SELECT * 
FROM networks
WHERE $1 << cidr;

-- name: ListNetworkUsers :many
SELECT u.*
FROM users u
JOIN network_members nm ON nm.user_id = u.id
WHERE nm.network_id = $1;

