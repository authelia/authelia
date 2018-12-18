
export interface UserInfo {
  username: string;
  password_hash: string;
  email: string;
  groups?: string[];
}

export type UserDatabaseConfiguration = UserInfo[];