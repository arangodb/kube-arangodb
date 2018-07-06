import React, { Component } from 'react';
import LogoutContext from '../auth/LogoutContext.js';
import DeploymentDetails from './DeploymentDetails.js';
import DeploymentList from './DeploymentList.js';
import { Header, Menu, Segment } from 'semantic-ui-react';
import { BrowserRouter as Router, Route, Link } from "react-router-dom";
import styled from 'react-emotion';

const StyledMenu = styled(Menu)`
  width: 15rem !important;
  @media (max-width: 768px) {
    width: 10rem !important;
  }
`;

const StyledContentBox = styled('div')`
  margin-left: 15rem;
  @media (max-width: 768px) {
    margin-left: 10rem;
  }
`;

const ListView = () => (
  <div>
    <Header dividing>
      ArangoDeployments
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
                  <Link to="/">Deployments</Link>
                </Menu.Item>
                <Menu.Item position="right" onClick={() => doLogout()}>
                  Logout
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
            {this.props["pod-info"]}
          </StyledContentBox>
        </div>
      </Router>
    );
  }
}

export default DeploymentOperator;
