import PortalReducer from './Portal';
import { StateType } from 'typesafe-actions';

function getReturnType<R> (f: (...args: any[]) => R): R {
  return null!;
}

const t = getReturnType(PortalReducer)

export type RootState = StateType<typeof t>;

export default PortalReducer;