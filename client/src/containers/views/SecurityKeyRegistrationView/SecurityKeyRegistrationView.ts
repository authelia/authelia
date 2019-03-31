import { connect } from 'react-redux';
import SecurityKeyRegistrationView from '../../../views/SecurityKeyRegistrationView/SecurityKeyRegistrationView';
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import * as U2fApi from "u2f-api";
import { Props } from '../../../views/SecurityKeyRegistrationView/SecurityKeyRegistrationView';
import { registerSecurityKey, registerSecurityKeyFailure, registerSecurityKeySuccess } from '../../../reducers/Portal/SecurityKeyRegistration/actions';
import AutheliaService from '../../../services/AutheliaService';

const mapStateToProps = (state: RootState) => ({
  deviceRegistered: state.securityKeyRegistration.success,
  error: state.securityKeyRegistration.error,
});

async function checkIdentity(token: string) {
  return fetch(`/api/secondfactor/u2f/identity/finish?token=${token}`, {
    method: 'POST',
  });
}

function fail(dispatch: Dispatch, err: Error) {
  dispatch(registerSecurityKeyFailure(err.message));
}

const mapDispatchToProps = (dispatch: Dispatch, ownProps: Props) => {
  return {
    onInit: async (token: string) => {
      try {
        dispatch(registerSecurityKey());
        await checkIdentity(token);
        const requestRegister = await AutheliaService.requestSecurityKeyRegistration();
        const registerResponse = await U2fApi.register(requestRegister, [], 60);
        await AutheliaService.completeSecurityKeyRegistration(registerResponse);
        dispatch(registerSecurityKeySuccess());
        setTimeout(() => {
          ownProps.history.push('/');
        }, 2000);
      } catch(err) {
        fail(dispatch, err);
      }
    },
    onBackClicked: () => {
      ownProps.history.push('/');
    }
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(SecurityKeyRegistrationView);