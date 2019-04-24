import { connect } from 'react-redux';
import SecondFactorTOTP, { StateProps, OwnProps } from "../../../components/SecondFactorTOTP/SecondFactorTOTP";
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import {
  oneTimePasswordVerification,
  oneTimePasswordVerificationFailure,
  oneTimePasswordVerificationSuccess
} from '../../../reducers/Portal/SecondFactor/actions';
import AutheliaService from '../../../services/AutheliaService';
import { push } from 'connected-react-router';
import FetchStateBehavior from '../../../behaviors/FetchStateBehavior';


const mapStateToProps = (state: RootState): StateProps => ({
  oneTimePasswordVerificationInProgress: state.secondFactor.oneTimePasswordVerificationLoading,
  oneTimePasswordVerificationError: state.secondFactor.oneTimePasswordVerificationError,
});

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
      try {
        dispatch(oneTimePasswordVerification());
        const response = await AutheliaService.verifyTotpToken(token, ownProps.redirectionUrl);
        dispatch(oneTimePasswordVerificationSuccess());
        if (response) {
          window.location.href = response.redirect;
          return;
        }
        await handleSuccess(dispatch);
      } catch (err) {
        console.error(err);
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