CREATE DATABASE graphql;

\c graphql;

CREATE TABLE users (
    username    char(30) CONSTRAINT firstkey PRIMARY KEY,
    admin       boolean,
    active      boolean
);

INSERT INTO users(username, admin, active) VALUES ('slomek',true,true);
