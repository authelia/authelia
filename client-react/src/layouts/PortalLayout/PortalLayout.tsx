import React, { Component } from "react";

import styles from "./PortalLayout.module.css"
import { Route, Switch, Redirect, RouterProps, RouteProps } from "react-router";

import { routes } from '../../routes/routes';
import { AUTHELIA_GITHUB_URL } from "../../constants";

interface Props extends RouterProps, RouteProps {}

export default class PortalLayout extends Component<Props> {

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
      <div className={styles.mainContent}>
        <div className={styles.frame}>
          <div className={styles.innerFrame}>
            <div className={styles.title}>
              {this.renderTitle()}
            </div>
            <div className={styles.content}>
              <Switch>
                {routes.map((r, key) => {
                  return <Route path={r.path} component={r.component} exact={true} key={key} />
                })}
                <Redirect to='/' />
              </Switch>
            </div>
          </div>
        </div>
        <div className={styles.footer}>
          <div>Powered by <a href={AUTHELIA_GITHUB_URL}>Authelia</a></div>
        </div>
      </div>
    )
  }
}