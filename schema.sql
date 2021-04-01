CREATE TABLE IF NOT EXISTS "users"
(
    "id"              integer PRIMARY KEY AUTOINCREMENT NOT NULL,
    "email"           varchar,
    "title"           varchar,
    "organization_id" integer

);

CREATE TABLE IF NOT EXISTS "expenses"
(
    "id"          integer PRIMARY KEY AUTOINCREMENT NOT NULL,
    "user_id"     integer,
    "amount"      integer,
    "description" varchar,
    CONSTRAINT "fk_expenses_users"
        FOREIGN KEY ("user_id")
            REFERENCES "users" ("id")
);

CREATE TABLE IF NOT EXISTS "organizations"
(
    "id"         integer PRIMARY KEY AUTOINCREMENT NOT NULL,
    "name"       varchar
);
