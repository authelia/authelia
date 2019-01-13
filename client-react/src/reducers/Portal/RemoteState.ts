import AuthenticationLevel from '../../types/AuthenticationLevel';

interface RemoteState {
  username: string;
  authentication_level: AuthenticationLevel;
}

export default RemoteState;