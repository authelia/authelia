import { connect } from 'react-redux';
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import { push } from 'connected-react-router';
import ForgotPasswordView from '../../../views/ForgotPasswordView/ForgotPasswordView';
import { forgotPasswordRequest, forgotPasswordSuccess, forgotPasswordFailure } from '../../../reducers/Portal/ForgotPassword/actions';
import AutheliaService from '../../../services/AutheliaService';

const mapStateToProps = (state: RootState) => ({
  disabled: state.forgotPassword.loading,
});

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    onPasswordResetRequested: async (username: string) => {
      try {
        dispatch(forgotPasswordRequest());
        await AutheliaService.initiatePasswordResetIdentityValidation(username);
        dispatch(forgotPasswordSuccess());
        await dispatch(push('/confirmation-sent'));
      } catch (err) {
        dispatch(forgotPasswordFailure(err.message));
      }
    },
    onCancelClicked: async () => {
      dispatch(push('/'));
    }
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(ForgotPasswordView);