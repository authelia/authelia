import { connect } from 'react-redux';
import SecondFactorTOTP, { StateProps, OwnProps } from "../../../components/SecondFactorTOTP/SecondFactorTOTP";
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import {
  oneTimePasswordVerification,
  oneTimePasswordVerificationFailure,
  oneTimePasswordVerificationSuccess
} from '../../../reducers/Portal/SecondFactor/actions';
import to from 'await-to-js';
import AutheliaService from '../../../services/AutheliaService';
import { push } from 'connected-react-router';
import FetchStateBehavior from '../../../behaviors/FetchStateBehavior';


const mapStateToProps = (state: RootState): StateProps => ({
  oneTimePasswordVerificationInProgress: state.secondFactor.oneTimePasswordVerificationLoading,
  oneTimePasswordVerificationError: state.secondFactor.oneTimePasswordVerificationError,
});

async function redirectIfPossible(dispatch: Dispatch, res: Response) {
  if (res.status === 204) return;

  const body = await res.json();
  if ('error' in body) {
    throw new Error(body['error']);
  }

  if ('redirect' in body) {
    window.location.href = body['redirect'];
    return;
  }
  return;
}

async function handleSuccess(dispatch: Dispatch, duration?: number) {
  async function handle() {
    await FetchStateBehavior(dispatch);
  }

  if (duration) {
    setTimeout(handle, duration);
  } else {
    await handle();
  }
}

const mapDispatchToProps = (dispatch: Dispatch, ownProps: OwnProps) => {
  return {
    onOneTimePasswordValidationRequested: async (token: string) => {
      let err, res;
      dispatch(oneTimePasswordVerification());
      [err, res] = await to(AutheliaService.verifyTotpToken(token, ownProps.redirectionUrl));
      if (err) {
        dispatch(oneTimePasswordVerificationFailure(err.message));
        throw err;
      }
      if (!res) {
        dispatch(oneTimePasswordVerificationFailure('No response'));
        throw 'No response';
      }

      try {
        await redirectIfPossible(dispatch, res);
        dispatch(oneTimePasswordVerificationSuccess());
        await handleSuccess(dispatch);
      } catch (err) {
        dispatch(oneTimePasswordVerificationFailure(err.message));
      }
    },
    onRegisterOneTimePasswordClicked: async () => {
      await AutheliaService.startTOTPRegistrationIdentityProcess();
      await dispatch(push('/confirmation-sent'));
    },
  }
}


export default connect(mapStateToProps, mapDispatchToProps)(SecondFactorTOTP);