import { ActionType, getType } from "typesafe-actions";
import * as Actions from './actions';
import { Secret } from "../../../views/OneTimePasswordRegistrationView/Secret";

type OneTimePasswordRegistrationAction = ActionType<typeof Actions>

export interface State {
  loading: boolean;
  error: string | null;
  secret: Secret | null;
}

let initialState: State = {
  loading: true,
  error: null,
  secret: null,
}

initialState = {
  secret: {
    base32_secret: 'PBSFWU2RM42HG3TNIRHUQMKSKVUW6NCNOBNFOLCFJZATS6CTI47A',
    otpauth_url: 'PBSFWU2RM42HG3TNIRHUQMKSKVUW6NCNOBNFOLCFJZATS6CTI47A',
  },
  error: null,
  loading: false,
}

export default (state = initialState, action: OneTimePasswordRegistrationAction) => {
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