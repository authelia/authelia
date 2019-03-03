import { connect } from 'react-redux';
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import u2fApi from 'u2f-api';
import to from 'await-to-js';
import {
  securityKeySignSuccess,
  securityKeySign,
  securityKeySignFailure,
  setSecurityKeySupported,
  oneTimePasswordVerification,
  oneTimePasswordVerificationFailure,
  oneTimePasswordVerificationSuccess
} from '../../../reducers/Portal/SecondFactor/actions';
import SecondFactorForm, { OwnProps, StateProps } from '../../../components/SecondFactorForm/SecondFactorForm';
import * as AutheliaService from '../../../services/AutheliaService';
import { push } from 'connected-react-router';
import fetchState from '../../../behaviors/FetchStateBehavior';
import LogoutBehavior from '../../../behaviors/LogoutBehavior';

const mapStateToProps = (state: RootState): StateProps => ({
  securityKeySupported: state.secondFactor.securityKeySupported,
  securityKeyVerified: state.secondFactor.securityKeySignSuccess || false,
  securityKeyError: state.secondFactor.error,

  oneTimePasswordVerificationInProgress: state.secondFactor.oneTimePasswordVerificationLoading,
  oneTimePasswordVerificationError: state.secondFactor.oneTimePasswordVerificationError,
});

async function triggerSecurityKeySigning(dispatch: Dispatch, redirectionUrl: string | null) {
  let err, result;
  dispatch(securityKeySign());
  [err, result] = await to(AutheliaService.requestSigning());
  if (err) {
    await dispatch(securityKeySignFailure(err.message));
    throw err;
  }

  if (!result) {
    await dispatch(securityKeySignFailure('No response'));
    throw 'No response';
  }

  [err, result] = await to(u2fApi.sign(result, 60));
  if (err) {
    await dispatch(securityKeySignFailure(err.message));
    throw err;
  }

  if (!result) {
    await dispatch(securityKeySignFailure('No response'));
    throw 'No response';
  }

  [err, result] = await to(AutheliaService.completeSecurityKeySigning(result, redirectionUrl));
  if (err) {
    await dispatch(securityKeySignFailure(err.message));
    throw err;
  }
  
  try {
    await redirectIfPossible(dispatch, result as Response);
    dispatch(securityKeySignSuccess());
    await handleSuccess(dispatch, 1000);
  } catch (err) {
    dispatch(securityKeySignFailure(err.message));
  }
}

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
    await fetchState(dispatch);
  }

  if (duration) {
    setTimeout(handle, duration);
  } else {
    await handle();
  }
}

const mapDispatchToProps = (dispatch: Dispatch, ownProps: OwnProps) => {
  return {
    onLogoutClicked: () => LogoutBehavior(dispatch),
    onRegisterSecurityKeyClicked: async () => {
      await AutheliaService.startU2FRegistrationIdentityProcess();
      await dispatch(push('/confirmation-sent'));
    },
    onRegisterOneTimePasswordClicked: async () => {
      await AutheliaService.startTOTPRegistrationIdentityProcess();
      await dispatch(push('/confirmation-sent'));
    },
    onInit: async () => {
      const isU2FSupported = await u2fApi.isSupported();
      if (isU2FSupported) {
        await dispatch(setSecurityKeySupported(true));
        await triggerSecurityKeySigning(dispatch, ownProps.redirectionUrl);
      }
    },
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
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(SecondFactorForm);