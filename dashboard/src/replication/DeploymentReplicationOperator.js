import { BrowserRouter as Router, Route, Link } from "react-router-dom";
import { Header, Menu, Message, Segment } from 'semantic-ui-react';
import React, { Component } from 'react';

import { StyledMenu, StyledContentBox } from '../style/style';
import DeploymentReplicationDetails from './DeploymentReplicationDetails';
import DeploymentReplicationList from './DeploymentReplicationList';
import LogoutContext from '../auth/LogoutContext';

const ListView = () => (
  <div>
    <Header dividing>
      ArangoDeploymentReplication resources
    </Header>
    <DeploymentReplicationList/>
  </div>
);

const DetailView = ({match}) => (
  <div>
    <Header dividing>
      ArangoDeploymentReplication {match.params.name}
    </Header>
    <DeploymentReplicationDetails name={match.params.name}/>
  </div>
);

class DeploymentReplicationOperator extends Component {
  render() {
    return (
      <Router>
        <div>
          <LogoutContext.Consumer>
            {doLogout => 
              <StyledMenu fixed="left" vertical>
                <Menu.Item>
                  <Menu.Header>Deployment Replication Operator</Menu.Header>
                  <Menu.Menu>
                    <Menu.Item>
                      <Link to="/">Deployment replications</Link>
                    </Menu.Item>
                    <Menu.Item position="right" onClick={() => doLogout()}>
                      Logout
                    </Menu.Item>
                  </Menu.Menu>
                </Menu.Item>
                {this.props.commonMenuItems}
              </StyledMenu>
            }
          </LogoutContext.Consumer>
          <StyledContentBox>
            <Segment basic clearing>
                <div>
                  <Route exact path="/" component={ListView} />
                  <Route path="/deployment-replication/:name" component={DetailView} />
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

export default DeploymentReplicationOperator;
