CREATE TABLE IF NOT EXISTS cluster (
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    provider TEXT NOT NULL,
    token TEXT,
    ip    TEXT,
    cluster_cidr TEXT,
    master_extra_args TEXT,
    workerExtraArgs TEXT,
    registry TEXT,
    dataStore TEXT,
    k3sVersion TEXT,
    k3sChannel TEXT,
    installScript TEXT,
    mirror TEXT,
    dockerMirror TEXT,
    dockerScript TEXT,
    network  TEXT,
    ui INTEGER,
    cloudControllerManager INTEGER,
    cluster INTEGER,
    options BLOB,
    UNIQUE (name, provider)
);

CREATE TABLE IF NOT EXISTS node (
    cluster_id INTEGER
    instance_id TEXT
    instance_status TEXT
    public_ip_addresses TEXT
    internal_ip_addresses TEXT
    eip_allocation_ids TEXT
    master INTEGER
);