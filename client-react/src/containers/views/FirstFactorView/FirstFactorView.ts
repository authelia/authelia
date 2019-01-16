import { connect } from 'react-redux';
import QueryString from 'query-string';
import FirstFactorView, { Props } from '../../../views/FirstFactorView/FirstFactorView';
import { Dispatch } from 'redux';
import { authenticateFailure, authenticateSuccess, authenticate } from '../../../reducers/Portal/FirstFactor/actions';
import { RootState } from '../../../reducers';


const mapStateToProps = (state: RootState) => ({});

function redirect2FA(props: Props) {
  if (!props.location) {
    props.history.push('/2fa');
    return;
  }
  const params = QueryString.parse(props.location.search);

  if ('rd' in params) {
    const rd = params['rd'] as string;
    props.history.push(`/2fa?rd=${rd}`);
    return;
  }
  props.history.push('/2fa');
}

function onAuthenticationRequested(dispatch: Dispatch, ownProps: Props) {
  return async (username: string, password: string) => {
    dispatch(authenticate());
    fetch('/api/firstfactor', {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: username,
        password: password,
      })
    }).then(async (res) => {
      const json = await res.json();
      if ('error' in json) {
        dispatch(authenticateFailure(json['error']));
        return;
      }
      dispatch(authenticateSuccess());
      redirect2FA(ownProps);
    });
  }
}

const mapDispatchToProps = (dispatch: Dispatch, ownProps: Props) => {
  return {
    onAuthenticationRequested: onAuthenticationRequested(dispatch, ownProps),
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(FirstFactorView);