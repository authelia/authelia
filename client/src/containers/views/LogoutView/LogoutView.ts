import { connect } from 'react-redux';
import { Dispatch } from 'redux';
import { RootState } from '../../../reducers';
import LogoutView, { DispatchProps } from '../../../views/LogoutView/LogoutView';
import LogoutBehavior from '../../../behaviors/LogoutBehavior';

const mapStateToProps = (state: RootState) => {
  return {};
}

const mapDispatchToProps = (dispatch: Dispatch): DispatchProps => {
  return {
    onInit: () => LogoutBehavior(dispatch),
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(LogoutView);