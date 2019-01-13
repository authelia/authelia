import RemoteState from '../../reducers/Portal/RemoteState';

export interface WithState {
  state: RemoteState | null;
  stateError: string | null;
  stateLoading: boolean;
}