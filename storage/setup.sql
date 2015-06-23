CREATE TABLE checks (
    id BIGSERIAL PRIMARY KEY,
    name text,
    type text,
    send_alerts text,
    backoff text,
    config text,
    triggers text
);

CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    name text,
    type text,
    config text
);
