

CREATE TABLE users (
                       id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4 (),
                       username varchar  NOT NULL ,
                       hashed_password varchar NOT NULL,
                       email varchar UNIQUE NOT NULL,
                       created_at timestamptz NOT NULL DEFAULT (now()),
                       updated_at timestamptz NOT NULL DEFAULT (now()),
                       deleted_at timestamptz
);


CREATE TABLE sessions (
                          id uuid PRIMARY KEY NOT NULL,
                          email varchar NOT NULL,
                          refresh_token varchar NOT NULL,
                          expires_at timestamptz NOT NULL,
                          created_at timestamptz NOT NULL DEFAULT (now()),
                          updated_at timestamptz NOT NULL DEFAULT (now()),
                          deleted_at timestamptz
);

ALTER TABLE sessions ADD CONSTRAINT fk_sessions_users FOREIGN KEY (email)
    REFERENCES users (email);


CREATE TABLE private_chats (
                               message_id bigint PRIMARY KEY NOT NULL,
                               message_from uuid NOT NULL ,
                               message_to uuid NOT NULL,
                               content text NOT NULL ,
                               created_at timestamptz NOT NULL DEFAULT (now()),
                               updated_at timestamptz NOT NULL DEFAULT (now()),
                               deleted_at timestamptz
);

ALTER TABLE private_chats ADD CONSTRAINT fk_private_chats_users_from FOREIGN KEY (message_from)
    REFERENCES users (id);


ALTER TABLE private_chats ADD CONSTRAINT fk_private_chats_users_to FOREIGN KEY (message_to)
    REFERENCES users (id);


CREATE TABLE group_chats (
                             channel_id bigint  NOT NULL,
                             message_id uuid NOT NULL,
                             user_id bigint NOT NULL,
                             content text NOT NULL ,
                             created_at timestamptz NOT NULL DEFAULT (now()),
                             updated_at timestamptz NOT NULL DEFAULT (now()),
                             deleted_at timestamptz
);

ALTER TABLE ONLY  group_chats ADD CONSTRAINT "ID_PKEY" PRIMARY KEY (channel_id,message_id);


ALTER TABLE group_chats ADD CONSTRAINT fk_group_chats_users_from FOREIGN KEY (message_id)
    REFERENCES users (id);


-- Contact Table
CREATE TABLE contacts (
                          id uuid PRIMARY KEY NOT NULL ,
                          contact_one uuid NOT NULL,
                          contact_two uuid NOT NULL,
                          created_at timestamptz NOT NULL DEFAULT (now()),
                          updated_at timestamptz NOT NULL DEFAULT (now()),
                          deleted_at timestamptz
);

ALTER TABLE contacts ADD CONSTRAINT fk_contacts_users_one FOREIGN KEY (contact_one)
    REFERENCES users (id);

ALTER TABLE contacts ADD CONSTRAINT fk_contacts_users_two FOREIGN KEY (contact_two)
    REFERENCES users (id);

