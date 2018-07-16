import { BrowserRouter as Router, Route, Link } from "react-router-dom";
import { Header, Menu, Message, Segment } from 'semantic-ui-react';
import React, { Component } from 'react';

import { StyledMenu, StyledContentBox } from '../style/style';
import DeploymentDetails from './DeploymentDetails';
import DeploymentList from './DeploymentList';
import LogoutContext from '../auth/LogoutContext';

const ListView = () => (
  <div>
    <Header dividing>
      ArangoDeployment resources
    </Header>
    <DeploymentList/>
  </div>
);

const DetailView = ({match}) => (
  <div>
    <Header dividing>
      ArangoDeployment {match.params.name}
    </Header>
    <DeploymentDetails name={match.params.name}/>
  </div>
);

class DeploymentOperator extends Component {
  render() {
    return (
      <Router>
        <div>
          <LogoutContext.Consumer>
            {doLogout => 
              <StyledMenu fixed="left" vertical>
                <Menu.Item>
                  <Menu.Header>Deployment Operator</Menu.Header>
                    <Menu.Menu>
                      <Menu.Item>
                        <Link to="/">Deployments</Link>
                      </Menu.Item>
                      <Menu.Item position="right" onClick={() => doLogout()}>
                        Logout
                      </Menu.Item>
                    </Menu.Menu>
                  {this.props.commonMenuItems}
                </Menu.Item>
              </StyledMenu>
            }
          </LogoutContext.Consumer>
          <StyledContentBox>
            <Segment basic clearing>
                <div>
                  <Route exact path="/" component={ListView} />
                  <Route path="/deployment/:name" component={DetailView} />
                </div>
            </Segment>
            {this.props.podInfoView}
            {(this.props.error) ? <Segment basic><Message error content={this.props.error}/></Segment> : null}
          </StyledContentBox>
        </div>
      </Router>
    );
  }
}

export default DeploymentOperator;
