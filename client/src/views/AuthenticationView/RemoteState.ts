import AuthenticationLevel from '../../types/AuthenticationLevel';

interface RemoteState {
  username: string;
  authentication_level: AuthenticationLevel;
  default_redirection_url: string;
  method: 'u2f' | 'totp'
}

export default RemoteState;