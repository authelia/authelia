CREATE TABLE IF NOT EXISTS user_action (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	user_elevated_session_id INTEGER NULL,
	performed DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    username VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    action VARCHAR(20) NOT NULL,
    data TEXT,
	CONSTRAINT user_action_user_elevated_session_id_fkey
		FOREIGN KEY (user_elevated_session_id)
			REFERENCES user_elevated_session (id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX user_action_username ON user_action (username);
