
import * as Actions from './actions';
import { ActionType, getType } from 'typesafe-actions';

export type FirstFactorAction = ActionType<typeof Actions>;

enum Result {
  NONE,
  SUCCESS,
  FAILURE,
}

interface FirstFactorState {
  lastResult: Result;
  loading: boolean;
  error: string | null;
  username: string;
  password: string;
}

const firstFactorInitialState: FirstFactorState = {
  lastResult: Result.NONE,
  loading: false,
  error: null,
  username: '',
  password: '',
}

export default (state = firstFactorInitialState, action: FirstFactorAction): FirstFactorState => {
  switch(action.type) {
    case getType(Actions.authenticate):
      return {
        ...state,
        loading: true,
        error: null
      };
    case getType(Actions.authenticateSuccess):
      return {
        ...state,
        lastResult: Result.SUCCESS,
        loading: false,
        error: null,
      };
    case getType(Actions.authenticateFailure):
      return {
        ...state,
        lastResult: Result.FAILURE,
        loading: false,
        error: action.payload,
      };
    case getType(Actions.setUsername):
      return {
        ...state,
        username: action.payload,
      }
    case getType(Actions.setPassword):
      return {
        ...state,
        password: action.payload,
      }
  }
  return state;
}