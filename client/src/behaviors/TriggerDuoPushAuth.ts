import { Dispatch } from "redux";
import AutheliaService from "../services/AutheliaService";
import { triggerDuoPushAuth, triggerDuoPushAuthSuccess, triggerDuoPushAuthFailure } from "../reducers/Portal/SecondFactor/actions";

export default async function(dispatch: Dispatch, redirectionUrl: string | null) {
  dispatch(triggerDuoPushAuth());
  try {
    const body = await AutheliaService.triggerDuoPush(redirectionUrl);
    dispatch(triggerDuoPushAuthSuccess());
    return body;
  } catch (err) {
    console.error(err);
    dispatch(triggerDuoPushAuthFailure(err.message))
  }
}