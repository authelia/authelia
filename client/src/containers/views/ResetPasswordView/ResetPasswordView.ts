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
      await AutheliaService.completePasswordResetIdentityValidation(token);
    },
    onPasswordResetRequested: async (newPassword: string) => {
      await AutheliaService.resetPassword(newPassword);
      await dispatch(push('/'));
    },
    onCancelClicked: async () => {
      await dispatch(push('/'));
    }
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(ResetPasswordView);