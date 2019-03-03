import { ActionType, getType } from "typesafe-actions";
import * as Actions from './actions';
import { Secret } from "../../../views/OneTimePasswordRegistrationView/Secret";

type OneTimePasswordRegistrationAction = ActionType<typeof Actions>

export interface OneTimePasswordRegistrationState {
  loading: boolean;
  error: string | null;
  secret: Secret | null;
}

let oneTimePasswordRegistrationInitialState: OneTimePasswordRegistrationState = {
  loading: true,
  error: null,
  secret: null,
}

export default (state = oneTimePasswordRegistrationInitialState, action: OneTimePasswordRegistrationAction): OneTimePasswordRegistrationState => {
  switch(action.type) {
    case getType(Actions.generateTotpSecret):
      return {
        loading: true,
        error: null,
        secret: null,
      };
    case getType(Actions.generateTotpSecretSuccess):
      return {
        ...state,
        loading: false,
        secret: action.payload,
      }
    case getType(Actions.generateTotpSecretFailure):
      return {
        ...state,
        loading: false,
        error: action.payload,
      }
  }
  return state;
}