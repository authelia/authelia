import { connect } from 'react-redux';
import OneTimePasswordRegistrationView from '../../../views/OneTimePasswordRegistrationView/OneTimePasswordRegistrationView';
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import { generateTotpSecret, generateTotpSecretSuccess, generateTotpSecretFailure } from '../../../reducers/Portal/OneTimePasswordRegistration/actions';
import { push } from 'connected-react-router';
import AutheliaService from '../../../services/AutheliaService';

const mapStateToProps = (state: RootState) => ({
  error: state.oneTimePasswordRegistration.error,
  secret: state.oneTimePasswordRegistration.secret,
});

async function tryGenerateTotpSecret(dispatch: Dispatch, token: string) {
  try {
    dispatch(generateTotpSecret());
    const res = await AutheliaService.completeOneTimePasswordRegistrationIdentityValidation(token);
    dispatch(generateTotpSecretSuccess(res));
  } catch (err) {
    dispatch(generateTotpSecretFailure(err.message));
  }
}

const mapDispatchToProps = (dispatch: Dispatch) => {
  let internalToken: string;
  return {
    onInit: async (token: string) => {
      internalToken = token;
      await tryGenerateTotpSecret(dispatch, internalToken);
    },
    onRetryClicked: async () => {
      await tryGenerateTotpSecret(dispatch, internalToken);
    },
    onCancelClicked: () => {
      dispatch(push('/'));
    },
    onLoginClicked: () => {
      dispatch(push('/'));
    }
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(OneTimePasswordRegistrationView);