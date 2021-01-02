import { getState } from "../services/State";
import { useRemoteCall } from "./RemoteCall";

export function useAutheliaState() {
    return useRemoteCall(getState, []);
}
