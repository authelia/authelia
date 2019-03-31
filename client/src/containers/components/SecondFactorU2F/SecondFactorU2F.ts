import { connect } from 'react-redux';
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import SecondFactorU2F, { StateProps, OwnProps } from '../../../components/SecondFactorU2F/SecondFactorU2F';
import AutheliaService from '../../../services/AutheliaService';
import { push } from 'connected-react-router';
import u2fApi from 'u2f-api';
import to from 'await-to-js';
import {
  securityKeySignSuccess,
  securityKeySign,
  securityKeySignFailure,
} from '../../../reducers/Portal/SecondFactor/actions';
import FetchStateBehavior from '../../../behaviors/FetchStateBehavior';


const mapStateToProps = (state: RootState): StateProps => ({
  securityKeyVerified: state.secondFactor.securityKeySignSuccess || false,
  securityKeyError: state.secondFactor.error,
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
    await redirectIfPossible(result as Response);
    dispatch(securityKeySignSuccess());
    await handleSuccess(dispatch, 1000);
  } catch (err) {
    dispatch(securityKeySignFailure(err.message));
  }
}

async function redirectIfPossible(res: Response) {
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
    onRegisterSecurityKeyClicked: async () => {
      await AutheliaService.startU2FRegistrationIdentityProcess();
      await dispatch(push('/confirmation-sent'));
    },
    onInit: async () => {
      await triggerSecurityKeySigning(dispatch, ownProps.redirectionUrl);
    },
  }
}


export default connect(mapStateToProps, mapDispatchToProps)(SecondFactorU2F);