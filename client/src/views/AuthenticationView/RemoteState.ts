import AuthenticationLevel from '../../types/AuthenticationLevel';

interface RemoteState {
  username: string;
  authentication_level: AuthenticationLevel;
  default_redirection_url: string;
}

export default RemoteState;