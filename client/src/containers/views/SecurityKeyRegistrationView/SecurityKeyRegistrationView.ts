import { connect } from 'react-redux';
import SecurityKeyRegistrationView from '../../../views/SecurityKeyRegistrationView/SecurityKeyRegistrationView';
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import U2fApi from "u2f-api";
import { Props } from '../../../views/SecurityKeyRegistrationView/SecurityKeyRegistrationView';
import { registerSecurityKey, registerSecurityKeyFailure, registerSecurityKeySuccess } from '../../../reducers/Portal/SecurityKeyRegistration/actions';
import AutheliaService from '../../../services/AutheliaService';

const mapStateToProps = (state: RootState) => ({
  deviceRegistered: state.securityKeyRegistration.success,
  error: state.securityKeyRegistration.error,
});

const mapDispatchToProps = (dispatch: Dispatch, ownProps: Props) => {
  return {
    onInit: async (token: string) => {
      try {
        dispatch(registerSecurityKey());
        const registerRequest = await AutheliaService.completeSecurityKeyRegistrationIdentityValidation(token);
        const registerRequests: U2fApi.RegisterRequest[] = [];
        for(var i in registerRequest.registerRequests) {
          const r = registerRequest.registerRequests[i];
          registerRequests.push({
            appId: registerRequest.appId,
            challenge: r.challenge,
            version: r.version,
          })
        }
        const registerResponse = await U2fApi.register(registerRequests, [], 60);
        await AutheliaService.completeSecurityKeyRegistration(registerResponse);
        dispatch(registerSecurityKeySuccess());
        setTimeout(() => {
          ownProps.history.push('/');
        }, 2000);
      } catch(err) {
        console.error(err);
        dispatch(registerSecurityKeyFailure(err.message));
      }
    },
    onBackClicked: () => {
      ownProps.history.push('/');
    }
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(SecurityKeyRegistrationView);