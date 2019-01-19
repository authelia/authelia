
import * as Actions from './actions';
import { ActionType, getType } from 'typesafe-actions';
import RemoteState from '../../../views/AuthenticationView/RemoteState';

export type Action = ActionType<typeof Actions>;

interface State {
  redirectionUrl : string | null;
  remoteState: RemoteState | null;
  remoteStateLoading: boolean;
  remoteStateError: string | null;
}

const initialState: State = {
  redirectionUrl: null,

  remoteState: null,
  remoteStateLoading: false,
  remoteStateError: null,
}

export default (state = initialState, action: Action): State => {
  switch(action.type) {
    case getType(Actions.fetchState):
      return {
        ...state,
        remoteState: null,
        remoteStateError: null,
        remoteStateLoading: true,
      };
    case getType(Actions.fetchStateSuccess):
      return {
        ...state,
        remoteState: action.payload,
        remoteStateLoading: false,
      };
    case getType(Actions.fetchStateFailure):
      return {
        ...state,
        remoteStateError: action.payload,
        remoteStateLoading: false,
      };
    case getType(Actions.setRedirectionUrl):
      return {
        ...state,
        redirectionUrl: action.payload,
      }
  }
  return state;
}