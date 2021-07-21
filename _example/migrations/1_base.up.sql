CREATE TABLE users(
    ID uuid NOT NULL PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    email VARCHAR(256) NOT NULL UNIQUE,
    hash VARCHAR(512),
    active BOOLEAN NOT NULL DEFAULT false,
    last_signed_in TIMESTAMP DEFAULT null,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    version INT NOT NULL
);

CREATE TABLE roles(
    ID VARCHAR(128) NOT NULL PRIMARY KEY,
    label varchar(128) NOT NULL,
    version INT NOT NULL
);

CREATE TABLE scopes(
    name varchar(128) NOT NULL PRIMARY KEY
);

CREATE TABLE user_roles(
    user_id uuid REFERENCES users(ID),
    role_id VARCHAR(128) REFERENCES roles(ID),
    CONSTRAINT ux_user_id_role_id UNIQUE(user_id, role_id)
);
CREATE INDEX idx_user_id_user_roles ON user_roles(user_id);
CREATE INDEX idx_role_id_user_roles ON user_roles(role_id);

CREATE TABLE role_scopes(
    role_id VARCHAR(128) REFERENCES roles(ID),
    scope_id VARCHAR(128) REFERENCES scopes(name)
);
CREATE INDEX idx_role_id_role_scopes ON role_scopes(role_id);
CREATE INDEX idx_scope_id_role_scopes ON role_scopes(scope_id);

ALTER TABLE user_roles
DROP CONSTRAINT user_roles_role_id_fkey,
ADD CONSTRAINT user_roles_role_id_fkey
    FOREIGN KEY (role_id)
    REFERENCES roles(id)
    ON DELETE CASCADE;

ALTER TABLE user_roles
DROP CONSTRAINT user_roles_user_id_fkey,
ADD CONSTRAINT user_roles_user_id_fkey
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE;

ALTER TABLE role_scopes
DROP CONSTRAINT role_scopes_role_id_fkey,
ADD CONSTRAINT role_scopes_role_id_fkey
    FOREIGN KEY (role_id)
    REFERENCES roles(id)
    ON DELETE CASCADE;

ALTER TABLE role_scopes
DROP CONSTRAINT role_scopes_scope_id_fkey,
ADD CONSTRAINT role_scopes_scope_id_fkey
    FOREIGN KEY (scope_id)
    REFERENCES scopes(name)
    ON DELETE CASCADE;


CREATE TABLE features(
    name VARCHAR(64) NOT NULL PRIMARY KEY
);

CREATE TABLE plans(
    ID uuid PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE,
    label VARCHAR(64) NOT NULL,
    price INTEGER NOT NULL default 0,
    public BOOLEAN NOT NULL DEFAULT false,
    recurring BOOLEAN NOT NULL DEFAULT false,
    expires TIMESTAMP DEFAULT null,
    frequency INTEGER DEFAULT null
);

CREATE TABLE plan_feature(
    plan_id uuid REFERENCES plans(ID) ON DELETE CASCADE,
    feature_id VARCHAR(64) REFERENCES features(name) ON DELETE CASCADE
);

CREATE TABLE subscriptions(
    ID uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(ID) ON DELETE CASCADE,
    plan_id uuid NOT NULL REFERENCES plans(ID) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    active BOOLEAN NOT NULL,
    multiplier REAL NOT NULL DEFAULT 1,
    status VARCHAR(32) NOT NULL,
    foreign_subscription VARCHAR(256) DEFAULT null
);
CREATE UNIQUE INDEX subscriptions_user_id_plan_id_unique ON subscriptions (user_id) WHERE active = true;

CREATE TABLE payment_statuses(
    name VARCHAR(64) NOT NULL PRIMARY KEY
);

CREATE TABLE payments(
    ID uuid PRIMARY KEY,
    amount INT NOT NULL,
    completed_at TIMESTAMP NOT NULL DEFAULT now(),
    status VARCHAR(64) NOT NULL,
    subscription_id uuid DEFAULT null REFERENCES subscriptions(ID),
    user_id uuid REFERENCES users(ID),
    CONSTRAINT fk_status
        FOREIGN KEY(status)
        REFERENCES payment_statuses(name)
        ON DELETE CASCADE
);
