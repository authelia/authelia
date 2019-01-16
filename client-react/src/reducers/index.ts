import PortalReducer from './Portal';
import { StateType } from 'typesafe-actions';

export type RootState = StateType<typeof PortalReducer>;

export default PortalReducer;