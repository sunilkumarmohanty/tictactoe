BEGIN;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE games (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    computer_mark CHAR(1) NOT NULL,
    board VARCHAR(9) NOT NULL,
    status varchar(7) NOT NULL
);

COMMIT;