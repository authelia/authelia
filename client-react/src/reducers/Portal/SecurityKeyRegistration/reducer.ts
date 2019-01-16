import { ActionType, getType } from "typesafe-actions";
import * as Actions from './actions';

type SecurityKeyRegistrationAction = ActionType<typeof Actions>

export interface State {
  error: string | null;
  success: boolean | null;
}

let initialState: State = {
  error: null,
  success: null,
}

export default (state = initialState, action: SecurityKeyRegistrationAction): State => {
  switch(action.type) {
    case getType(Actions.registerSecurityKey):
      return {
        success: null,
        error: null,
      };
    case getType(Actions.registerSecurityKeySuccess):
      return {
        ...state,
        success: true,
      };
    case getType(Actions.registerSecurityKeyFailure):
      return {
        ...state,
        success: false,
        error: action.payload,
      };
  }
  return state;
}