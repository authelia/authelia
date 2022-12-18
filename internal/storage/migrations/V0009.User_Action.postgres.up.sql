CREATE TABLE IF NOT EXISTS user_action (
    id SERIAL CONSTRAINT user_action_pkey PRIMARY KEY,
	user_elevated_session_id INTEGER NULL,
	performed TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    username VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    action VARCHAR(20) NOT NULL,
    data TEXT
);

CREATE INDEX user_action_username ON user_action (username);

ALTER TABLE user_action
	ADD CONSTRAINT user_action_user_elevated_session_id_fkey
		FOREIGN KEY (user_elevated_session_id)
			REFERENCES user_elevated_session (id) ON UPDATE CASCADE ON DELETE CASCADE;
