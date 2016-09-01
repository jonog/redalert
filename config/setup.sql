CREATE TABLE checks (
    id text PRIMARY KEY NOT NULL CONSTRAINT non_empty CHECK(length(id)>0),
    name text NOT NULL,
    type text NOT NULL,
    send_alerts text,
    backoff text,
    config text,
    assertions text
);

CREATE TABLE notifications (
    id text PRIMARY KEY NOT NULL CONSTRAINT non_empty CHECK(length(id)>0),
    name text NOT NULL UNIQUE,
    type text NOT NULL,
    config text
);
