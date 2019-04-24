import { connect } from 'react-redux';
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import { push } from 'connected-react-router';
import ResetPasswordView, { StateProps } from '../../../views/ResetPasswordView/ResetPasswordView';
import AutheliaService from '../../../services/AutheliaService';

const mapStateToProps = (state: RootState): StateProps => ({
  disabled: state.resetPassword.loading,
});

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    onInit: async (token: string) => {
      try {
        await AutheliaService.completePasswordResetIdentityValidation(token);
      } catch (err) {
        console.error(err);
      }
    },
    onPasswordResetRequested: async (newPassword: string) => {
      try {
        await AutheliaService.resetPassword(newPassword);
        await dispatch(push('/'));
      } catch (err) {
        console.error(err);
      }
    },
    onCancelClicked: async () => {
      await dispatch(push('/'));
    }
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(ResetPasswordView);