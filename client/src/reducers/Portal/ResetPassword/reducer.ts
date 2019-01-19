
import * as Actions from './actions';
import { ActionType, getType } from 'typesafe-actions';

export type Action = ActionType<typeof Actions>;


interface State {
  loading: boolean;
  success: boolean | null;
  error: string | null;
}

const initialState: State = {
  loading: false,
  success: null,
  error: null,
}

export default (state = initialState, action: Action): State => {
  switch(action.type) {
    case getType(Actions.resetPasswordRequest):
      return {
        ...state,
        loading: true,
        error: null
      };
    case getType(Actions.resetPasswordSuccess):
      return {
        ...state,
        success: true,
        loading: false,
        error: null,
      };
    case getType(Actions.resetPasswordFailure):
      return {
        ...state,
        success: false,
        loading: false,
        error: action.payload,
      };
  }
  return state;
}