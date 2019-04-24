import { connect } from 'react-redux';
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import SecondFactorU2F, { StateProps, OwnProps } from '../../../components/SecondFactorU2F/SecondFactorU2F';
import AutheliaService from '../../../services/AutheliaService';
import { push } from 'connected-react-router';
import u2fApi from 'u2f-api';
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
  dispatch(securityKeySign());
  const signRequest = await AutheliaService.requestSigning();
  const signRequests: u2fApi.SignRequest[] = [];
  for (var i in signRequest.registeredKeys) {
    const r = signRequest.registeredKeys[i];
    signRequests.push({
      appId: signRequest.appId,
      challenge: signRequest.challenge,
      keyHandle: r.keyHandle,
      version: r.version,
    })
  }
  const signResponse = await u2fApi.sign(signRequests, 60);
  const response = await AutheliaService.completeSecurityKeySigning(signResponse, redirectionUrl);
  dispatch(securityKeySignSuccess());

  if (response) {
    window.location.href = response.redirect;
    return;
  }
  await handleSuccess(dispatch, 1000);
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
      try {
        await triggerSecurityKeySigning(dispatch, ownProps.redirectionUrl);
      } catch (err) {
        console.error(err);
        await dispatch(securityKeySignFailure(err.message));
      }
    },
  }
}


export default connect(mapStateToProps, mapDispatchToProps)(SecondFactorU2F);