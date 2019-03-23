import { Dispatch } from "redux";
import { setPreferedMethod, setPreferedMethodSuccess, setPreferedMethodFailure } from "../reducers/Portal/SecondFactor/actions";
import AutheliaService from "../services/AutheliaService";
import Method2FA from "../types/Method2FA";

export default async function(dispatch: Dispatch, method: Method2FA) {
  dispatch(setPreferedMethod());
  try {
    await AutheliaService.setPrefered2faMethod(method);
    dispatch(setPreferedMethodSuccess());
  } catch (err) {
    dispatch(setPreferedMethodFailure(err.message))
  }
}