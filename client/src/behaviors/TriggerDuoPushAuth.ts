import { Dispatch } from "redux";
import AutheliaService from "../services/AutheliaService";
import { triggerDuoPushAuth, triggerDuoPushAuthSuccess, triggerDuoPushAuthFailure } from "../reducers/Portal/SecondFactor/actions";

export default async function(dispatch: Dispatch, redirectionUrl: string | null) {
  dispatch(triggerDuoPushAuth());
  try {
    const res = await AutheliaService.triggerDuoPush(redirectionUrl);
    const body = await res.json();
    if ('error' in body) {
      throw new Error(body['error']);
    }
    dispatch(triggerDuoPushAuthSuccess());
    return body;
  } catch (err) {
    dispatch(triggerDuoPushAuthFailure(err.message))
  }
}