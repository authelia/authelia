
import * as Actions from './actions';
import { ActionType, getType, StateType } from 'typesafe-actions';
import Method2FA from '../../../types/Method2FA';

export type SecondFactorAction = ActionType<typeof Actions>;

interface SecondFactorState {
  logoutLoading: boolean;
  logoutSuccess: boolean | null;
  error: string | null;

  userAnotherMethod: boolean;

  getAvailableMethodsLoading: boolean;
  getAvailableMethodResponse: Method2FA[] | null;
  getAvailableMethodError: string | null;

  preferedMethodLoading: boolean;
  preferedMethodError: string | null;
  preferedMethod: Method2FA | null;

  setPreferedMethodLoading: boolean;
  setPreferedMethodError: string | null;
  setPreferedMethodSuccess: boolean | null;

  securityKeySupported: boolean;
  securityKeySignLoading: boolean;
  securityKeySignSuccess: boolean | null;

  oneTimePasswordVerificationLoading: boolean,
  oneTimePasswordVerificationSuccess: boolean | null,
  oneTimePasswordVerificationError: string | null,

  duoPushVerificationLoading: boolean;
  duoPushVerificationSuccess: boolean | null;
  duoPushVerificationError: string | null;
}

const secondFactorInitialState: SecondFactorState = {
  logoutLoading: false,
  logoutSuccess: null,
  error: null,

  userAnotherMethod: false,

  getAvailableMethodsLoading: false,
  getAvailableMethodResponse: null,
  getAvailableMethodError: null,

  preferedMethod: null,
  preferedMethodError: null,
  preferedMethodLoading: false,

  setPreferedMethodLoading: false,
  setPreferedMethodError: null,
  setPreferedMethodSuccess: null,

  securityKeySupported: false,
  securityKeySignLoading: false,
  securityKeySignSuccess: null,

  oneTimePasswordVerificationLoading: false,
  oneTimePasswordVerificationError: null,
  oneTimePasswordVerificationSuccess: null,

  duoPushVerificationLoading: false,
  duoPushVerificationSuccess: null,
  duoPushVerificationError: null,
}

export type PortalState = StateType<SecondFactorState>;

export default (state = secondFactorInitialState, action: SecondFactorAction): SecondFactorState => {
  switch(action.type) {
    case getType(Actions.logout):
      return {
        ...state,
        logoutLoading: true,
        logoutSuccess: null,
        error: null,
      };
    case getType(Actions.logoutSuccess):
      return {
        ...state,
        logoutLoading: false,
        logoutSuccess: true,
      };
    case getType(Actions.logoutFailure):
      return {
        ...state,
        logoutLoading: false,
        error: action.payload,
      }
      
    case getType(Actions.securityKeySign):
      return {
        ...state,
        securityKeySignLoading: true,
        securityKeySignSuccess: false,
      };
    case getType(Actions.securityKeySignSuccess):
      return {
        ...state,
        securityKeySignLoading: false,
        securityKeySignSuccess: true,
      };
    case getType(Actions.securityKeySignFailure):
      return {
        ...state,
        securityKeySignLoading: false,
        securityKeySignSuccess: false,
      };

    case getType(Actions.setSecurityKeySupported):
      return {
        ...state,
        securityKeySupported: action.payload,
      };

    case getType(Actions.oneTimePasswordVerification):
      return {
        ...state,
        oneTimePasswordVerificationLoading: true,
        oneTimePasswordVerificationError: null,
      }
    case getType(Actions.oneTimePasswordVerificationSuccess):
      return {
        ...state,
        oneTimePasswordVerificationLoading: false,
        oneTimePasswordVerificationSuccess: true,
      }
    case getType(Actions.oneTimePasswordVerificationFailure):
      return {
        ...state,
        oneTimePasswordVerificationLoading: false,
        oneTimePasswordVerificationError: action.payload,
      }
    case getType(Actions.getPreferedMethod):
      return {
        ...state,
        preferedMethodLoading: true,
        preferedMethod: null,
        preferedMethodError: null,
      }
    case getType(Actions.getPreferedMethodSuccess):
      return {
        ...state,
        preferedMethodLoading: false,
        preferedMethod: action.payload,
      }
    case getType(Actions.getPreferedMethodFailure):
      return {
        ...state,
        preferedMethodLoading: false,
        preferedMethodError: action.payload,
      }
    case getType(Actions.setPreferedMethod):
      return {
        ...state,
        setPreferedMethodLoading: true,
        setPreferedMethodSuccess: null,
        preferedMethodError: null,
      }
    case getType(Actions.getPreferedMethodSuccess):
      return {
        ...state,
        setPreferedMethodLoading: false,
        setPreferedMethodSuccess: true,
      }
    case getType(Actions.getPreferedMethodFailure):
      return {
        ...state,
        setPreferedMethodLoading: false,
        setPreferedMethodError: action.payload,
      }
    case getType(Actions.setUseAnotherMethod):
      return {
        ...state,
        userAnotherMethod: action.payload,
      }
    case getType(Actions.triggerDuoPushAuth):
      return {
        ...state,
        duoPushVerificationLoading: true,
        duoPushVerificationError: null,
        duoPushVerificationSuccess: null,
      }
    case getType(Actions.triggerDuoPushAuthSuccess):
      return {
        ...state,
        duoPushVerificationLoading: false,
        duoPushVerificationSuccess: true,
      }
    case getType(Actions.triggerDuoPushAuthFailure):
      return {
        ...state,
        duoPushVerificationLoading: false,
        duoPushVerificationError: action.payload,
      }

    case getType(Actions.getPreferedMethod):
      return {
        ...state,
        getAvailableMethodsLoading: true,
        getAvailableMethodResponse: null,
        getAvailableMethodError: null,
      }
    case getType(Actions.getAvailbleMethodsSuccess):
      return {
        ...state,
        getAvailableMethodsLoading: false,
        getAvailableMethodResponse: action.payload,
      }
    case getType(Actions.getAvailbleMethodsFailure):
      return {
        ...state,
        getAvailableMethodsLoading: false,
        getAvailableMethodError: action.payload,
      }
  }
  return state;
}