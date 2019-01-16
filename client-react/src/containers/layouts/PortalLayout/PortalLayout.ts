import { connect } from 'react-redux';
import PortalLayout from '../../../layouts/PortalLayout/PortalLayout';
import { RootState } from '../../../reducers';

const mapStateToProps = (state: RootState) => ({
  authenticationLevel: (state.firstFactor.remoteState) ? state.firstFactor.remoteState.authentication_level : 0,
});

export default connect(mapStateToProps)(PortalLayout);