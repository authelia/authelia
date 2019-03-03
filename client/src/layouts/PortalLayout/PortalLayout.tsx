import React, { Component } from "react";

import { Route, Switch, Redirect, RouterProps, RouteProps } from "react-router";

import { routes } from '../../routes/routes';
import { AUTHELIA_GITHUB_URL } from "../../constants";

import styles from '../../assets/scss/layouts/PortalLayout/PortalLayout.module.scss';

interface Props extends RouterProps, RouteProps {}

class PortalLayout extends Component<Props> {
  private renderTitle() {
    if (!this.props.location) return;

    for (let i in routes) {
      const route = routes[i];
      if (route.path && route.path.indexOf(this.props.location.pathname) > -1) {
        return route.title.toUpperCase();
      }
    }
    return;
  }


  render() {
    return (
      <div className={styles.main}>
        <div className={styles.mainContent}>
          <div className={styles.title}>
            {this.renderTitle()}
          </div>
          <div className={styles.frame}>
            <div className={styles.innerFrame}>
              <Switch>
                {routes.map((r, key) => {
                  return <Route path={r.path} component={r.component} exact={true} key={key} />
                })}
                <Redirect to='/' />
              </Switch>
            </div>
          </div>
          <div className={styles.footer}>
            <div><a href={AUTHELIA_GITHUB_URL}>Powered by Authelia</a></div>
          </div>
        </div>
      </div>
    )
  }
}

export default PortalLayout;